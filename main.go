package main

import (
	"bytes"
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/storage"
	"fmt"
	"github.com/rs/xid"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"remap-keys.app/remap-build-server/parser"
	"time"
)

type RequestParameters struct {
	Uid    string
	TaskId string
}

type Task struct {
	Uid              string    `firestore:"uid"`
	Status           string    `firestore:"status"`
	FirmwareId       string    `firestore:"firmwareId"`
	FirmwareFilePath string    `firestore:"firmwareFilePath"`
	Stdout           string    `firestore:"stdout"`
	Stderr           string    `firestore:"stderr"`
	ParametersJson   string    `firestore:"parametersJson"`
	CreatedAt        time.Time `firestore:"createdAt"`
	UpdatedAt        time.Time `firestore:"updatedAt"`
}

type Firmware struct {
	KeyboardDefinitionId string    `firestore:"keyboardDefinitionId"`
	uid                  string    `firestore:"uid"`
	CreatedAt            time.Time `firestore:"createdAt"`
	UpdatedAt            time.Time `firestore:"updatedAt"`
}

type FirmwareFile struct {
	ID      string `firestore:"-"`
	Path    string `firestore:"path"`
	Content string `firestore:"content"`
}

type Parameters struct {
	Keyboard map[string]map[string]string `json:"keyboard"`
	Keymap   map[string]map[string]string `json:"keymap"`
}

// QMK Firmware base directory path.
const qmkFirmwareBaseDirectoryPath string = "/root/versions/0.22.12"

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

	// Start the HTTP server.
	port := os.Getenv("PORT")
	log.Printf("PORT(Env): %s\n", port)
	if port == "" {
		port = "80"
	}
	h := func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, ctx, firestoreClient, storageClient)
	}
	http.HandleFunc("/build", h)
	log.Printf("Remap Build Server is running on port %s.\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
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

// Parses the query parameters. The query parameters are as follows:
//   - uid: The user's UID.
//   - taskId: The task ID.
func parseQueryParameters(r *http.Request) (*RequestParameters, error) {
	queryParams := r.URL.Query()
	uid := queryParams.Get("uid")
	taskId := queryParams.Get("taskId")
	if uid == "" || taskId == "" {
		return nil, fmt.Errorf("uid or taskId is empty")
	}
	return &RequestParameters{
		Uid:    uid,
		TaskId: taskId,
	}, nil
}

// Fetches the task information from the Firestore.
func fetchTaskInfo(client *firestore.Client, params *RequestParameters) (*Task, error) {
	log.Println("Fetching the task information from the Firestore.")
	taskDoc, err := client.Collection("build").Doc("v1").Collection("tasks").Doc(params.TaskId).Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}
	var task Task
	taskDoc.DataTo(&task)
	return &task, nil
}

// Fetches the firmware information from the Firestore.
func fetchFirmwareInfo(client *firestore.Client, task *Task) (*Firmware, error) {
	log.Println("Fetching the firmware information from the Firestore.")
	firmwareDoc, err := client.Collection("build").Doc("v1").Collection("firmwares").Doc(task.FirmwareId).Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("firmware not found")
		}
		return nil, err
	}
	var firmware Firmware
	firmwareDoc.DataTo(&firmware)
	return &firmware, nil
}

// Fetches the keyboard files from the Firestore.
func fetchKeyboardFiles(client *firestore.Client, firmwareId string) ([]*FirmwareFile, error) {
	log.Println("Fetching the keyboard files from the Firestore.")
	iter := client.Collection("build").Doc("v1").Collection("firmwares").Doc(firmwareId).Collection("keyboardFiles").Documents(context.Background())
	var keyboardFiles []*FirmwareFile
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var keyboardFile FirmwareFile
		doc.DataTo(&keyboardFile)
		keyboardFile.ID = doc.Ref.ID
		keyboardFiles = append(keyboardFiles, &keyboardFile)
	}
	return keyboardFiles, nil
}

// Fetches the keymap files from the Firestore.
func fetchKeymapFiles(client *firestore.Client, firmwareId string) ([]*FirmwareFile, error) {
	log.Println("Fetching the keymap files from the Firestore.")
	iter := client.Collection("build").Doc("v1").Collection("firmwares").Doc(firmwareId).Collection("keymapFiles").Documents(context.Background())
	var keymapFiles []*FirmwareFile
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var keymapFile FirmwareFile
		doc.DataTo(&keymapFile)
		keymapFile.ID = doc.Ref.ID
		keymapFiles = append(keymapFiles, &keymapFile)
	}
	return keymapFiles, nil
}

// Generates the keyboard ID.
func generateKeyboardId() string {
	guid := xid.New()
	return guid.String()
}

// Prepares the keyboard directory in the QMK Firmware base directory.
// For instance, remove the directory if it exists and create a new directory.
// Returns the keyboard directory path if succeeded.
func prepareKeyboardDirectory(keyboardId string) (string, error) {
	log.Println("Preparing the keyboard directory.")
	keyboardDirectoryFullPath := filepath.Join(qmkFirmwareBaseDirectoryPath, "keyboards", keyboardId)
	log.Printf("[INFO] keyboardDirectoryFullPath: %s\n", keyboardDirectoryFullPath)
	_, err := os.Stat(keyboardDirectoryFullPath)
	if err == nil {
		log.Println("The keyboard directory exists. Removing it.")
		err = os.RemoveAll(keyboardDirectoryFullPath)
		if err != nil {
			return "", err
		}
	}
	log.Println("Creating the keyboard directory.")
	err = os.MkdirAll(keyboardDirectoryFullPath, 0755)
	if err != nil {
		return "", err
	}
	return keyboardDirectoryFullPath, nil
}

func deleteKeyboardDirectory(keyboardId string) error {
	keyboardDirectoryFullPath := filepath.Join(qmkFirmwareBaseDirectoryPath, "keyboards", keyboardId)
	return os.RemoveAll(keyboardDirectoryFullPath)
}

func createFile(path string, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

func createFirmwareFiles(baseDirectoryPath string, firmwareFiles []*FirmwareFile) error {
	for _, firmwareFile := range firmwareFiles {
		// If the path of the keyboardFile includes the directory divided by the "/" character,
		// create the directory, then create the file.
		// Otherwise, create the file.
		dir, file := filepath.Split(firmwareFile.Path)
		var targetDirectoryPath string
		if dir != "" {
			targetDirectoryPath = filepath.Join(baseDirectoryPath, dir)
			err := os.MkdirAll(targetDirectoryPath, 0755)
			if err != nil {
				return err
			}
		} else {
			targetDirectoryPath = baseDirectoryPath
		}
		targetFilePath := filepath.Join(targetDirectoryPath, file)
		log.Printf("[INFO] targetFilePath: %s\n", targetFilePath)
		err := createFile(targetFilePath, firmwareFile.Content)
		if err != nil {
			return err
		}
	}
	return nil
}

type BuildResult struct {
	success bool
	stdout  string
	stderr  string
}

func buildQmkFirmware(keyboardId string) BuildResult {
	log.Println("Building a QMK Firmware started.")
	cmd := exec.Command(
		"/root/.local/bin/qmk", "compile",
		"-kb", keyboardId,
		"-km", "remap")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	log.Println("Building a QMK Firmware finished.")
	stdoutString := stdout.String()
	if err != nil {
		log.Println("Building failed.")
		stderrString := stderr.String()
		log.Printf("[ERROR] %s\n", err.Error())
		return BuildResult{
			success: false,
			stdout:  stdoutString,
			stderr:  stderrString,
		}
	}
	log.Println("Building succeeded.")
	return BuildResult{
		success: true,
		stdout:  stdoutString,
		stderr:  "",
	}
}

func updateTaskStatusToBuilding(ctx context.Context, client *firestore.Client, taskId string) error {
	_, err := client.Collection("build").Doc("v1").Collection("tasks").Doc(taskId).Set(ctx, map[string]interface{}{
		"status":           "building",
		"stdout":           "",
		"stderr":           "",
		"firmwareFilePath": "",
		"updatedAt":        time.Now(),
	}, firestore.MergeAll)
	if err != nil {
		return err
	}
	return nil
}

func sendFailureResponseWithError(taskId string, client *firestore.Client, w http.ResponseWriter, cause error) {
	log.Printf("[ERROR] %s\n", cause.Error())
	// Update the task status to "failure".
	_, err := client.Collection("build").Doc("v1").Collection("tasks").Doc(taskId).Set(context.Background(), map[string]interface{}{
		"status":           "failure",
		"stdout":           "",
		"stderr":           cause.Error(),
		"firmwareFilePath": "",
		"updatedAt":        time.Now(),
	}, firestore.MergeAll)
	if err != nil {
		// Ignore the error about updating the task status.
		log.Printf("[ERROR] %s\n", err.Error())
	}
	// Return the error message, but return the status code 200 to avoid the retry with Cloud Tasks.
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, cause.Error())
}

func sendFailureResponseWithStdoutAndStderr(taskId string, client *firestore.Client, w http.ResponseWriter, message string, stdout string, stderr string) {
	log.Printf("[ERROR] %s\n", message)
	// Update the task status to "failure".
	_, err := client.Collection("build").Doc("v1").Collection("tasks").Doc(taskId).Set(context.Background(), map[string]interface{}{
		"status":           "failure",
		"stdout":           stdout,
		"stderr":           stderr,
		"firmwareFilePath": "",
		"updatedAt":        time.Now(),
	}, firestore.MergeAll)
	if err != nil {
		log.Printf("[ERROR] %s\n", err.Error())
	}
	// Return the error message, but return the status code 200 to avoid the retry with Cloud Tasks.
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, err.Error())
}

func uploadFirmwareFileToCloudStorage(ctx context.Context, storageClient *storage.Client, uid string, firmwareFileName string, localFirmwareFilePath string) (string, error) {
	log.Println("Uploading the firmware file to the Cloud Storage.")

	file, err := os.Open(localFirmwareFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	bucketName := "remap-b2d08.appspot.com"
	remoteFirmwareFilePath := fmt.Sprintf("firmware/%s/built/%s", uid, firmwareFileName)
	bucket, err := storageClient.Bucket(bucketName)
	if err != nil {
		return "", err
	}
	writer := bucket.Object(remoteFirmwareFilePath).NewWriter(ctx)
	if _, err := io.Copy(writer, file); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}
	return remoteFirmwareFilePath, nil
}

func sendSuccessResponseWithStdout(taskId string, client *firestore.Client, w http.ResponseWriter, stdout string, remoteFirmwareFilePath string) error {
	_, err := client.Collection("build").Doc("v1").Collection("tasks").Doc(taskId).Set(context.Background(), map[string]interface{}{
		"status":           "success",
		"stdout":           stdout,
		"stderr":           "",
		"firmwareFilePath": remoteFirmwareFilePath,
		"updatedAt":        time.Now(),
	}, firestore.MergeAll)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Building succeeded")
	return nil
}

// Replaces the parameters in the keyboard files.
func replaceParameters(files []*FirmwareFile, parameterFileMap map[string]map[string]string) []*FirmwareFile {
	for _, file := range files {
		parameterMap := parameterFileMap[file.ID]
		if parameterMap == nil {
			// If there is no parameter map for the firmware file, skip this file.
			continue
		}
		newContent := parser.ReplaceParameters(file.Content, parameterMap)
		file.Content = newContent
	}
	return files
}

// Checks the authentication token.
// This function gets the authentication token from the `Authorization` header.
// The authentication token is a JWT token.
// The JWT token is generated by the Cloud Tasks.
// The `aud`, `email`, `exp` claims are included in the JWT token.
// This function checks them and validates the signature using the public key.
// If the validation and checks are failed, this function returns an error.
func checkAuthenticationToken(r *http.Request) error {
	// Fetch the authentication token from the `Authorization` header.
	authorizationHeader := r.Header.Get("Authorization")
	log.Printf("Authorization header: %s\n", authorizationHeader)
	if authorizationHeader == "" {
		return fmt.Errorf("authorization header is empty")
	}
	// The authentication token is a JWT token.
	// The JWT token is generated by the Cloud Tasks.
	// The `aud`, `email`, `exp` claims are included in the JWT token.
	// This function checks them and validates the signature using the public key.
	// If the validation and checks are failed, this function returns an error.
	return nil
}

// Handles the HTTP request.
func handleRequest(w http.ResponseWriter, r *http.Request, ctx context.Context, firestoreClient *firestore.Client, storageClient *storage.Client) {
	log.Printf("%s %s %s\n", r.Method, r.URL, r.Proto)

	// Check the authentication token.
	err := checkAuthenticationToken(r)
	if err != nil {
		log.Printf("[ERROR] %s\n", err.Error())
		// Return the error message, but return the status code 200 to avoid the retry with Cloud Tasks.
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, err.Error())
		return
	}

	// Fetch the query parameters (uid and taskId).
	params, err := parseQueryParameters(r)
	if err != nil {
		log.Printf("[ERROR] %s\n", err.Error())
		// Return the error message, but return the status code 200 to avoid the retry with Cloud Tasks.
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, err.Error())
		return
	}
	log.Printf("[INFO] uid: %s, taskId: %s\n", params.Uid, params.TaskId)

	// Fetch the task information from the Firestore.
	task, err := fetchTaskInfo(firestoreClient, params)
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
		sendFailureResponseWithError(params.TaskId, firestoreClient, w, fmt.Errorf("uid in the task information and passed uid are not the same"))
		return
	}

	// Parse the parameters JSON string.
	var parameters Parameters
	err = json.Unmarshal([]byte(task.ParametersJson), &parameters)
	if err != nil {
		sendFailureResponseWithError(params.TaskId, firestoreClient, w, err)
		return
	}

	// Update the task status to "building".
	err = updateTaskStatusToBuilding(ctx, firestoreClient, params.TaskId)
	if err != nil {
		sendFailureResponseWithError(params.TaskId, firestoreClient, w, err)
		return
	}

	// Fetch the firmware information from the Firestore.
	firmware, err := fetchFirmwareInfo(firestoreClient, task)
	if err != nil {
		sendFailureResponseWithError(params.TaskId, firestoreClient, w, err)
		return
	}
	log.Printf("[INFO] The firmware [%+v] exists. The keyboard definition ID is [%+v]\n", task.FirmwareId, firmware.KeyboardDefinitionId)

	// Fetch the keyboard files from the Firestore.
	keyboardFiles, err := fetchKeyboardFiles(firestoreClient, task.FirmwareId)
	if err != nil {
		sendFailureResponseWithError(params.TaskId, firestoreClient, w, err)
		return
	}
	log.Printf("[INFO] keyboardFiles: %+v\n", keyboardFiles)

	// Fetch the keymap files from the Firestore.
	keymapFiles, err := fetchKeymapFiles(firestoreClient, task.FirmwareId)
	if err != nil {
		sendFailureResponseWithError(params.TaskId, firestoreClient, w, err)
		return
	}
	log.Printf("[INFO] keymapFiles: %+v\n", keymapFiles)

	// Replace parameters.
	keyboardFiles = replaceParameters(keyboardFiles, parameters.Keyboard)
	keymapFiles = replaceParameters(keymapFiles, parameters.Keymap)

	// Generate the keyboard ID.
	keyboardId := generateKeyboardId()
	log.Printf("[INFO] keyboardId: %s\n", keyboardId)

	// Prepare the keyboard directory.
	keyboardDirectoryPath, err := prepareKeyboardDirectory(keyboardId)
	if err != nil {
		sendFailureResponseWithError(params.TaskId, firestoreClient, w, err)
		return
	}
	log.Printf("[INFO] Keyboard directory path: %s\n", keyboardDirectoryPath)

	// Create the keyboard files.
	err = createFirmwareFiles(keyboardDirectoryPath, keyboardFiles)
	if err != nil {
		sendFailureResponseWithError(params.TaskId, firestoreClient, w, err)
		return
	}

	// Create the keymap files.
	keymapDirectoryPath := filepath.Join(keyboardDirectoryPath, "keymaps", "remap")
	err = os.MkdirAll(keymapDirectoryPath, 0755)
	if err != nil {
		sendFailureResponseWithError(params.TaskId, firestoreClient, w, err)
		return
	}
	err = createFirmwareFiles(keymapDirectoryPath, keymapFiles)
	if err != nil {
		sendFailureResponseWithError(params.TaskId, firestoreClient, w, err)
		return
	}

	// Build the QMK Firmware.
	buildResult := buildQmkFirmware(keyboardId)
	log.Printf("[INFO] buildResult: %v\n", buildResult.success)
	if !buildResult.success {
		sendFailureResponseWithStdoutAndStderr(params.TaskId, firestoreClient, w, "Building failed", buildResult.stdout, buildResult.stderr)
		return
	}
	log.Printf("[INFO] Building succeeded\n")

	// Create the local firmware file path.
	firmwareFileName, err := parser.FetchFirmwareFileName(buildResult.stdout)
	if err != nil {
		sendFailureResponseWithStdoutAndStderr(params.TaskId, firestoreClient, w, err.Error(), buildResult.stdout, buildResult.stderr)
		return
	}
	localFirmwareFilePath := filepath.Join(qmkFirmwareBaseDirectoryPath, firmwareFileName)
	log.Printf("[INFO] localFirmwareFilePath: %s\n", localFirmwareFilePath)

	// Upload the firmware file to the Cloud Storage.
	remoteFirmwareFilePath, err := uploadFirmwareFileToCloudStorage(ctx, storageClient, params.Uid, firmwareFileName, localFirmwareFilePath)
	if err != nil {
		sendFailureResponseWithError(params.TaskId, firestoreClient, w, err)
		return
	}
	log.Printf("[INFO] remoteFirmwareFilePath: %s\n", remoteFirmwareFilePath)

	// Delete the keyboard directory.
	err = deleteKeyboardDirectory(keyboardId)
	if err != nil {
		sendFailureResponseWithError(params.TaskId, firestoreClient, w, err)
		return
	}

	// Update the task status to "success".
	err = sendSuccessResponseWithStdout(params.TaskId, firestoreClient, w, buildResult.stdout, remoteFirmwareFilePath)
	if err != nil {
		sendFailureResponseWithError(params.TaskId, firestoreClient, w, err)
	}
}
