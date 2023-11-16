package database

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"remap-keys.app/remap-build-server/common"
	"time"
)

// FetchTaskInfo fetches the task information from the Firestore.
func FetchTaskInfo(client *firestore.Client, params *common.RequestParameters) (*common.Task, error) {
	log.Println("Fetching the task information from the Firestore.")
	taskDoc, err := client.Collection("build").Doc("v1").Collection("tasks").Doc(params.TaskId).Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}
	var task common.Task
	taskDoc.DataTo(&task)
	return &task, nil
}

// FetchFirmwareInfo fetches the firmware information from the Firestore.
func FetchFirmwareInfo(client *firestore.Client, task *common.Task) (*common.Firmware, error) {
	log.Println("Fetching the firmware information from the Firestore.")
	firmwareDoc, err := client.Collection("build").Doc("v1").Collection("firmwares").Doc(task.FirmwareId).Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("firmware not found")
		}
		return nil, err
	}
	var firmware common.Firmware
	firmwareDoc.DataTo(&firmware)
	return &firmware, nil
}

// FetchKeyboardFiles fetches the keyboard files from the Firestore.
func FetchKeyboardFiles(client *firestore.Client, firmwareId string) ([]*common.FirmwareFile, error) {
	log.Println("Fetching the keyboard files from the Firestore.")
	iter := client.Collection("build").Doc("v1").Collection("firmwares").Doc(firmwareId).Collection("keyboardFiles").Documents(context.Background())
	var keyboardFiles []*common.FirmwareFile
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var keyboardFile common.FirmwareFile
		doc.DataTo(&keyboardFile)
		keyboardFile.ID = doc.Ref.ID
		keyboardFiles = append(keyboardFiles, &keyboardFile)
	}
	return keyboardFiles, nil
}

// FetchKeymapFiles fetches the keymap files from the Firestore.
func FetchKeymapFiles(client *firestore.Client, firmwareId string) ([]*common.FirmwareFile, error) {
	log.Println("Fetching the keymap files from the Firestore.")
	iter := client.Collection("build").Doc("v1").Collection("firmwares").Doc(firmwareId).Collection("keymapFiles").Documents(context.Background())
	var keymapFiles []*common.FirmwareFile
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var keymapFile common.FirmwareFile
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

func FetchCertificate(ctx context.Context, client *firestore.Client, key string) (*common.Certificate, error) {
	iter := client.Collection("certificates").Where("domain", "==", key).Documents(ctx)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			log.Printf("[Error] Failed to fetch the certificate from the Firestore: %v", err)
			return nil, err
		}
		var certificate common.Certificate
		doc.DataTo(&certificate)
		certificate.ID = doc.Ref.ID
		log.Printf("[Info] Found the certificate from the Firestore: %v", certificate.ID)
		return &certificate, nil
	}
	log.Printf("[Info] Certificate not found: %v", key)
	return nil, nil
}

func SaveCertificate(ctx context.Context, client *firestore.Client, key string, data []byte) error {
	certificate, err := FetchCertificate(ctx, client, key)
	if err != nil {
		log.Printf("[Error] Failed to fetch the certificate from the Firestore: %v", err)
		return err
	}
	log.Printf("[Info] Saving the certificate to the Firestore: %v", key)
	if certificate == nil {
		_, _, err := client.Collection("certificates").Add(ctx, common.Certificate{Domain: key, Data: data})
		if err != nil {
			log.Printf("[Error] Failed to save the certificate to the Firestore: %v", err)
			return err
		}
	} else {
		_, err := client.Collection("certificates").Doc(certificate.ID).Set(ctx, common.Certificate{Domain: key, Data: data})
		if err != nil {
			log.Printf("[Error] Failed to save the certificate to the Firestore: %v", err)
			return err
		}
	}
	log.Printf("[Info] Saved the certificate to the Firestore: %v", key)
	return nil
}

func DeleteCertificate(ctx context.Context, client *firestore.Client, key string) error {
	certificate, err := FetchCertificate(ctx, client, key)
	if err != nil {
		log.Printf("[Error] Failed to fetch the certificate from the Firestore: %v", err)
		return err
	}
	if certificate == nil {
		return fmt.Errorf("certificate not found")
	}
	_, err = client.Collection("certificates").Doc(certificate.ID).Delete(ctx)
	if err != nil {
		log.Printf("[Error] Failed to delete the certificate from the Firestore: %v", err)
		return err
	}
	log.Printf("[Info] Deleted the certificate from the Firestore: %v", key)
	return nil
}
