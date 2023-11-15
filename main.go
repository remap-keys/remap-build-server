package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/storage"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"remap-keys.app/remap-build-server/auth"
	"remap-keys.app/remap-build-server/build"
	"remap-keys.app/remap-build-server/database"
	"remap-keys.app/remap-build-server/parameter"
	"remap-keys.app/remap-build-server/web"
	"time"
)

type Parameters struct {
	Keyboard map[string]map[string]string `json:"keyboard"`
	Keymap   map[string]map[string]string `json:"keymap"`
}

func main() {
	// Prepare the Firestore firestoreClient.
	ctx := context.Background()
	app := createFirebaseApp(ctx)
	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer firestoreClient.Close()
	storageClient, err := app.Storage(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	certCache := web.NewFirestoreCertCache(firestoreClient)
	certManager := autocert.Manager{
		Cache:      certCache,
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("build.remap-keys.app"),
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, ctx, firestoreClient, storageClient)
	}
	http.HandleFunc("/build", h)

	tlsConfig := &tls.Config{
		Rand:           rand.Reader,
		Time:           time.Now,
		NextProtos:     []string{http2.NextProtoTLS, "http/1.1"},
		MinVersion:     tls.VersionTLS12,
		GetCertificate: certManager.GetCertificate,
	}

	server := &http.Server{
		Addr:      ":https",
		TLSConfig: tlsConfig,
	}
	go http.ListenAndServe(":http", certManager.HTTPHandler(nil))
	log.Fatal(server.ListenAndServeTLS("", ""))
}

func createFirebaseApp(ctx context.Context) *firebase.App {
	//sa := option.WithCredentialsFile("service-account-remap-b2d08-70b4596e8a05.json")
	//app, err := firebase.NewApp(ctx, nil, sa)
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalln(err)
	}
	return app
}

func sendFailureResponseWithError(ctx context.Context, taskId string, client *firestore.Client, w http.ResponseWriter, cause error) {
	log.Printf("[ERROR] %s\n", cause.Error())
	// Update the task status to "failure".
	err := database.UpdateTask(ctx, client, taskId, "failure", "", cause.Error(), "")
	if err != nil {
		// Ignore the error about updating the task status.
		log.Printf("[ERROR] %s\n", err.Error())
	}
	// Return the error message, but return the status code 200 to avoid the retry with Cloud Tasks.
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, cause.Error())
}

func sendFailureResponseWithStdoutAndStderr(ctx context.Context, taskId string, client *firestore.Client, w http.ResponseWriter, message string, stdout string, stderr string) {
	log.Printf("[ERROR] %s\n", message)
	// Update the task status to "failure".
	err := database.UpdateTask(ctx, client, taskId, "failure", stdout, stderr, "")
	if err != nil {
		log.Printf("[ERROR] %s\n", err.Error())
	}
	// Return the error message, but return the status code 200 to avoid the retry with Cloud Tasks.
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, message)
}

func sendSuccessResponseWithStdout(ctx context.Context, taskId string, client *firestore.Client, w http.ResponseWriter, stdout string, remoteFirmwareFilePath string) error {
	err := database.UpdateTask(ctx, client, taskId, "success", stdout, "", remoteFirmwareFilePath)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Building succeeded")
	return nil
}

// Handles the HTTP request.
func handleRequest(w http.ResponseWriter, r *http.Request, ctx context.Context, firestoreClient *firestore.Client, storageClient *storage.Client) {
	log.Printf("%s %s %s\n", r.Method, r.URL, r.Proto)

	// Fetch the query parameters (uid and taskId).
	params, err := web.ParseQueryParameters(r)
	if err != nil {
		log.Printf("[ERROR] %s\n", err.Error())
		// Return the error message, but return the status code 200 to avoid the retry with Cloud Tasks.
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, err.Error())
		return
	}
	log.Printf("[INFO] uid: %s, taskId: %s\n", params.Uid, params.TaskId)

	// Fetch the task information from the Firestore.
	task, err := database.FetchTaskInfo(firestoreClient, params)
	if err != nil {
		log.Printf("[ERROR] %s\n", err.Error())
		// Return the error message, but return the status code 200 to avoid the retry with Cloud Tasks.
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, err.Error())
		return
	}
	log.Printf("[INFO] The task [%+v] exists\n", params.TaskId)

	// Check whether the uid in the task information and passed uid are the same.
	if task.Uid != params.Uid {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, fmt.Errorf("uid in the task information and passed uid are not the same"))
		return
	}

	// Check the authentication token.
	err = auth.CheckAuthenticationToken(r)
	if err != nil {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, err)
		return
	}

	// Parse the parameters JSON string.
	var parameters Parameters
	err = json.Unmarshal([]byte(task.ParametersJson), &parameters)
	if err != nil {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, err)
		return
	}

	// Update the task status to "building".
	err = database.UpdateTaskStatusToBuilding(ctx, firestoreClient, params.TaskId)
	if err != nil {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, err)
		return
	}

	// Fetch the firmware information from the Firestore.
	firmware, err := database.FetchFirmwareInfo(firestoreClient, task)
	if err != nil {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, err)
		return
	}
	log.Printf("[INFO] The firmware [%+v] exists. The keyboard definition ID is [%+v]\n", task.FirmwareId, firmware.KeyboardDefinitionId)

	// Check whether the firmware is enabled.
	if !firmware.Enabled {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, fmt.Errorf("the firmware is not enabled"))
		return
	}

	// Fetch the keyboard files from the Firestore.
	keyboardFiles, err := database.FetchKeyboardFiles(firestoreClient, task.FirmwareId)
	if err != nil {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, err)
		return
	}
	log.Printf("[INFO] keyboardFiles: %+v\n", keyboardFiles)

	// Fetch the keymap files from the Firestore.
	keymapFiles, err := database.FetchKeymapFiles(firestoreClient, task.FirmwareId)
	if err != nil {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, err)
		return
	}
	log.Printf("[INFO] keymapFiles: %+v\n", keymapFiles)

	// Replace parameters.
	keyboardFiles = parameter.ReplaceParameters(keyboardFiles, parameters.Keyboard)
	keymapFiles = parameter.ReplaceParameters(keymapFiles, parameters.Keymap)

	// Generate the keyboard ID.
	keyboardId := build.GenerateKeyboardId()
	log.Printf("[INFO] keyboardId: %s\n", keyboardId)

	// Prepare the keyboard directory.
	keyboardDirectoryPath, err := build.PrepareKeyboardDirectory(keyboardId, firmware.QmkFirmwareVersion)
	if err != nil {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, err)
		return
	}
	log.Printf("[INFO] Keyboard directory path: %s\n", keyboardDirectoryPath)

	// Delete the keyboard directory after the function returns.
	defer func() {
		// Delete the keyboard directory.
		err = build.DeleteKeyboardDirectory(keyboardId, firmware.QmkFirmwareVersion)
		if err != nil {
			log.Printf("[ERROR] %s\n", err.Error())
		}
		log.Printf("[INFO] Deleted the keyboard directory: %s\n", keyboardDirectoryPath)
	}()

	// Create the keyboard files.
	err = build.CreateFirmwareFiles(keyboardDirectoryPath, keyboardFiles)
	if err != nil {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, err)
		return
	}

	// Create the keymap files.
	keymapDirectoryPath := filepath.Join(keyboardDirectoryPath, "keymaps", "remap")
	err = os.MkdirAll(keymapDirectoryPath, 0755)
	if err != nil {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, err)
		return
	}
	err = build.CreateFirmwareFiles(keymapDirectoryPath, keymapFiles)
	if err != nil {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, err)
		return
	}

	// Build the QMK Firmware.
	buildResult := build.BuildQmkFirmware(keyboardId, firmware.QmkFirmwareVersion)
	log.Printf("[INFO] buildResult: %v\n", buildResult.Success)
	if !buildResult.Success {
		sendFailureResponseWithStdoutAndStderr(ctx, params.TaskId, firestoreClient, w, "Building failed", buildResult.Stdout, buildResult.Stderr)
		return
	}
	log.Printf("[INFO] Building succeeded\n")

	// Create the local firmware file path.
	firmwareFileName, err := parameter.FetchFirmwareFileName(buildResult.Stdout)
	if err != nil {
		sendFailureResponseWithStdoutAndStderr(ctx, params.TaskId, firestoreClient, w, err.Error(), buildResult.Stdout, buildResult.Stderr)
		return
	}
	localFirmwareFilePath := filepath.Join(
		build.QmkFirmwareBaseDirectoryPath+firmware.QmkFirmwareVersion, firmwareFileName)
	log.Printf("[INFO] localFirmwareFilePath: %s\n", localFirmwareFilePath)

	// Upload the firmware file to the Cloud Storage.
	remoteFirmwareFilePath, err := database.UploadFirmwareFileToCloudStorage(ctx, storageClient, params.Uid, firmwareFileName, localFirmwareFilePath)
	if err != nil {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, err)
		return
	}
	log.Printf("[INFO] remoteFirmwareFilePath: %s\n", remoteFirmwareFilePath)

	// Update the task status to "success".
	err = sendSuccessResponseWithStdout(ctx, params.TaskId, firestoreClient, w, buildResult.Stdout, remoteFirmwareFilePath)
	if err != nil {
		sendFailureResponseWithError(ctx, params.TaskId, firestoreClient, w, err)
	}
}
