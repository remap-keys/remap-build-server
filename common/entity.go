package common

import (
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
	KeyboardDefinitionId  string    `firestore:"keyboardDefinitionId"`
	Uid                   string    `firestore:"uid"`
	Enabled               bool      `firestore:"enabled"`
	QmkFirmwareVersion    string    `firestore:"qmkFirmwareVersion"`
	KeyboardDirectoryName string    `firestore:"keyboardDirectoryName"`
	CreatedAt             time.Time `firestore:"createdAt"`
	UpdatedAt             time.Time `firestore:"updatedAt"`
}

type FirmwareFile struct {
	ID      string `firestore:"-"`
	Path    string `firestore:"path"`
	Content string `firestore:"content"`
}

type Certificate struct {
	ID     string `firestore:"-"`
	Domain string `firestore:"domain"`
	Data   []byte `firestore:"data"`
}

type RequestParameters struct {
	Uid    string
	TaskId string
}

type ParametersJsonVersion1 struct {
	Keyboard map[string]map[string]string `json:"keyboard"`
	Keymap   map[string]map[string]string `json:"keymap"`
}

type ParametersJson struct {
	Version  int8                      `json:"version"`
	Keyboard map[string]ParameterValue `json:"keyboard"`
	Keymap   map[string]ParameterValue `json:"keymap"`
}

type ParameterValue struct {
	Type       string            `json:"type"`
	Parameters map[string]string `json:"parameters"`
	Code       string            `json:"code"`
}
