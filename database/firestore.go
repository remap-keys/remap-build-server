package database

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"remap-keys.app/remap-build-server/web"
	"time"
)

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
	Uid                  string    `firestore:"uid"`
	Enabled              bool      `firestore:"enabled"`
	QmkFirmwareVersion   string    `firestore:"qmkFirmwareVersion"`
	CreatedAt            time.Time `firestore:"createdAt"`
	UpdatedAt            time.Time `firestore:"updatedAt"`
}

type FirmwareFile struct {
	ID      string `firestore:"-"`
	Path    string `firestore:"path"`
	Content string `firestore:"content"`
}

// FetchTaskInfo fetches the task information from the Firestore.
func FetchTaskInfo(client *firestore.Client, params *web.RequestParameters) (*Task, error) {
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

// FetchFirmwareInfo fetches the firmware information from the Firestore.
func FetchFirmwareInfo(client *firestore.Client, task *Task) (*Firmware, error) {
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

// FetchKeyboardFiles fetches the keyboard files from the Firestore.
func FetchKeyboardFiles(client *firestore.Client, firmwareId string) ([]*FirmwareFile, error) {
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

// FetchKeymapFiles fetches the keymap files from the Firestore.
func FetchKeymapFiles(client *firestore.Client, firmwareId string) ([]*FirmwareFile, error) {
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

// UpdateTaskStatusToBuilding updates the task status to "building".
func UpdateTaskStatusToBuilding(ctx context.Context, client *firestore.Client, taskId string) error {
	return UpdateTask(ctx, client, taskId, "building", "", "", "")
}

func UpdateTask(ctx context.Context, client *firestore.Client, taskId string, status string, stdout string, stderr string, firmwareFilePath string) error {
	_, err := client.Collection("build").Doc("v1").Collection("tasks").Doc(taskId).Set(ctx, map[string]interface{}{
		"status":           status,
		"stdout":           stdout,
		"stderr":           stderr,
		"firmwareFilePath": firmwareFilePath,
		"updatedAt":        time.Now(),
	}, firestore.MergeAll)
	return err
}
