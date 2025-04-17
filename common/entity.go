package common

import (
	"time"
)

type Task struct {
	Uid              string    `firestore:"uid"`
	Status           string    `firestore:"status"`
	FirmwareId       string    `firestore:"firmwareId"`
	ProjectId        string    `firestore:"projectId"`
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

type WorkbenchProject struct {
	Name                  string    `firestore:"name"`
	QmkFirmwareVersion    string    `firestore:"qmkFirmwareVersion"`
	Uid                   string    `firestore:"uid"`
	KeyboardDirectoryName string    `firestore:"keyboardDirectoryName"`
	CreatedAt             time.Time `firestore:"createdAt"`
	UpdatedAt             time.Time `firestore:"updatedAt"`
}

type BuildableFile interface {
	GetPath() string
	GetContent() string
}

type FirmwareFile struct {
	ID      string `firestore:"-"`
	Path    string `firestore:"path"`
	Content string `firestore:"content"`
}

func (f FirmwareFile) GetPath() string {
	return f.Path
}

func (f FirmwareFile) GetContent() string {
	return f.Content
}

type WorkbenchProjectFile struct {
	ID        string    `firestore:"-"`
	Path      string    `firestore:"path"`
	Content   string    `firestore:"code"`
	FileType  string    `firestore:"fileType"`
	CreatedAt time.Time `firestore:"createdAt"`
	UpdatedAt time.Time `firestore:"updatedAt"`
}

func (w WorkbenchProjectFile) GetPath() string {
	return w.Path
}

func (w WorkbenchProjectFile) GetContent() string {
	return w.Content
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
	Version  int8                       `json:"version"`
	Keyboard map[string]*ParameterValue `json:"keyboard"`
	Keymap   map[string]*ParameterValue `json:"keymap"`
}

type ParameterValue struct {
	Type       string            `json:"type"`
	Parameters map[string]string `json:"parameters"`
	Code       string            `json:"code"`
}
