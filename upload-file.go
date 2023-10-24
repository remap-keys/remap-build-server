package main

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"flag"
	"google.golang.org/api/option"
	"io"
	"log"
	"os"
)

func createFirestoreClient(ctx context.Context) *firestore.Client {
	sa := option.WithCredentialsFile("service-account-remap-b2d08-70b4596e8a05.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	return client
}

type Parameters struct {
	FirmwareId string
	FileType   string
	FileName   string
	FilePath   string
}

func parseCommandArguments() *Parameters {
	var (
		firmwareId = flag.String("f", "", "The firmware ID.")
		fileType   = flag.String("t", "", "The file type.")
		fileName   = flag.String("n", "", "The file name.")
		filePath   = flag.String("p", "", "The file path.")
	)
	flag.Parse()

	if *firmwareId == "" {
		log.Fatalln("The firmware ID is required.")
	}
	if *fileType == "" {
		log.Fatalln("The file type is required.")
	}
	if *fileType != "keyboard" && *fileType != "keymap" {
		log.Fatalln("The file type must be either firmware or keymap.")
	}
	if *fileName == "" {
		log.Fatalln("The file name is required.")
	}
	if *filePath == "" {
		log.Fatalln("The file path is required.")
	}

	return &Parameters{
		FirmwareId: *firmwareId,
		FileType:   *fileType,
		FileName:   *fileName,
		FilePath:   *filePath,
	}
}

func readFileContent(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	return string(b)
}

func uploadFileToFirestore(ctx context.Context, client *firestore.Client, params *Parameters, content string) {
	var subCollectionName string
	if params.FileType == "keyboard" {
		subCollectionName = "keyboardFiles"
	} else {
		subCollectionName = "keymapFiles"
	}
	log.Printf("Sub collection name: %s\n", subCollectionName)
	log.Printf("Firmware ID: %s\n", params.FirmwareId)
	_, _, err := client.Collection("build").Doc("v1").Collection("firmwares").Doc(params.FirmwareId).Collection(subCollectionName).Add(ctx, map[string]interface{}{
		"path":    params.FileName,
		"content": content,
	})
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	ctx := context.Background()
	client := createFirestoreClient(ctx)
	defer client.Close()
	log.Println("Firestore client created.")
	params := parseCommandArguments()
	log.Println("Command arguments parsed.")
	content := readFileContent(params.FilePath)
	log.Println("File content read.")
	uploadFileToFirestore(ctx, client, params, content)
	log.Println("File uploaded to Firestore.")
}
