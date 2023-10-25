package main

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"flag"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	KeyboardId string
	FileType   string
	FileName   string
	FilePath   string
}

func parseCommandArguments() *Parameters {
	var (
		keyboardId = flag.String("k", "", "The keyboard ID.")
		fileType   = flag.String("t", "", "The file type.")
		fileName   = flag.String("n", "", "The file name.")
		filePath   = flag.String("p", "", "The file path.")
	)
	flag.Parse()

	if *keyboardId == "" {
		log.Fatalln("The keyboard ID is required.")
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
		KeyboardId: *keyboardId,
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
	log.Printf("Firmware ID: %s\n", params.KeyboardId)
	_, _, err := client.Collection("build").Doc("v1").Collection("firmwares").Doc(params.KeyboardId).Collection(subCollectionName).Add(ctx, map[string]interface{}{
		"path":    params.FileName,
		"content": content,
	})
	if err != nil {
		log.Fatalln(err)
	}
}

func createFirmwareDocument(ctx context.Context, client *firestore.Client, params *Parameters, keyboardDefinition *KeyboardDefinition) {
	firmwares := client.Collection("build").Doc("v1").Collection("firmwares")
	_, err := firmwares.Doc(params.KeyboardId).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			log.Printf("Firmware document not found: %s\n", params.KeyboardId)
		} else {
			log.Printf("Firmware document already exists: %s\n", params.KeyboardId)
			return
		}
	}
	_, err = firmwares.Doc(params.KeyboardId).Set(ctx, map[string]interface{}{
		"keyboardDefinitionId": params.KeyboardId,
		"uid":                  keyboardDefinition.AuthorUid,
		"createdAt":            firestore.ServerTimestamp,
		"updatedAt":            firestore.ServerTimestamp,
	})
	if err != nil {
		log.Fatalln(err)
	}
}

type KeyboardDefinition struct {
	AuthorUid string `firestore:"author_uid"`
}

func fetchKeyboardDefinitionDocument(ctx context.Context, client *firestore.Client, params *Parameters) *KeyboardDefinition {
	definitions := client.Collection("keyboards").Doc("v2").Collection("definitions")
	ref, err := definitions.Doc(params.KeyboardId).Get(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	var keyboardDefinition KeyboardDefinition
	ref.DataTo(&keyboardDefinition)
	log.Printf("Keyboard definition: %v\n", keyboardDefinition)
	return &keyboardDefinition
}

func main() {
	ctx := context.Background()
	client := createFirestoreClient(ctx)
	defer client.Close()
	log.Println("Firestore client created.")
	params := parseCommandArguments()
	log.Println("Command arguments parsed.")
	keyboardDefinition := fetchKeyboardDefinitionDocument(ctx, client, params)
	log.Println("Keyboard keyboardDefinition fetched.")
	createFirmwareDocument(ctx, client, params, keyboardDefinition)
	log.Println("Firmware document created.")
	content := readFileContent(params.FilePath)
	log.Println("File content read.")
	uploadFileToFirestore(ctx, client, params, content)
	log.Println("File uploaded to Firestore.")
}
