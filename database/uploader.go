package database

import (
	"context"
	"firebase.google.com/go/storage"
	"fmt"
	"io"
	"log"
	"os"
)

// UploadFirmwareFileToCloudStorage uploads the firmware file to the Cloud Storage.
func UploadFirmwareFileToCloudStorage(ctx context.Context, storageClient *storage.Client, uid string, firmwareFileName string, localFirmwareFilePath string) (string, error) {
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
