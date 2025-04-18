package build

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/xid"
	"remap-keys.app/remap-build-server/common"
)

// QmkFirmwareBaseDirectoryPath is QMK Firmware base directory path.
const QmkFirmwareBaseDirectoryPath string = "/root/versions/"

// BuildResult represents the result of the build.
type BuildResult struct {
	Success bool
	Stdout  string
	Stderr  string
}

// GenerateKeyboardId generates the keyboard ID.
// If the passed keyboard directory name is not empty string, use it.
// Otherwise, generate the keyboard ID.
func GenerateKeyboardId(keyboardDirectoryName string) string {
	if keyboardDirectoryName != "" {
		return keyboardDirectoryName
	}
	guid := xid.New()
	return guid.String()
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

// CreateFiles creates the files.
func CreateFiles(baseDirectoryPath string, buildableFiles []common.BuildableFile) error {
	for _, buildableFile := range buildableFiles {
		// If the path of the keyboardFile includes the directory divided by the "/" character,
		// create the directory, then create the file.
		// Otherwise, create the file.
		dir, file := filepath.Split(buildableFile.GetPath())
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
		err := createFile(targetFilePath, buildableFile.GetContent())
		if err != nil {
			return err
		}
	}
	return nil
}

// BuildQmkFirmware builds a QMK Firmware.
func BuildQmkFirmware(keyboardId string, qmkFirmwareVersion string) BuildResult {
	log.Println("Building a QMK Firmware started.")
	cmd := exec.Command(
		"/root/.local/bin/qmk", "compile",
		"-kb", keyboardId,
		"-km", "remap")
	cmd.Dir = "/root/versions/" + qmkFirmwareVersion
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "OPT_DEFS=-DBUILD_ON_REMAP")
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
			Success: false,
			Stdout:  stdoutString,
			Stderr:  stderrString,
		}
	}
	log.Println("Building succeeded.")
	return BuildResult{
		Success: true,
		Stdout:  stdoutString,
		Stderr:  "",
	}
}

// DeleteKeyboardDirectory deletes the keyboard directory.
func DeleteKeyboardDirectory(keyboardId string, qmkFirmwareVersion string) error {
	keyboardDirectoryFullPath := filepath.Join(
		QmkFirmwareBaseDirectoryPath+qmkFirmwareVersion, "keyboards", keyboardId)
	return os.RemoveAll(keyboardDirectoryFullPath)
}

// PrepareKeyboardDirectory prepares the keyboard directory in the QMK Firmware base directory.
// For instance, remove the directory if it exists and create a new directory.
// Returns the keyboard directory path if succeeded.
func PrepareKeyboardDirectory(keyboardId string, qmkFirmwareVersion string) (string, error) {
	log.Println("Preparing the keyboard directory.")
	keyboardDirectoryFullPath := filepath.Join(
		QmkFirmwareBaseDirectoryPath+qmkFirmwareVersion, "keyboards", keyboardId)
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

// CreateFirmwareFileNameWithTimestamp creates a firmware file name with timestamp.
// This function does the following:
//  1. Create the timestamp string based on the epoch.
//  2. Append the timestamp string after the firmware file name with the "_" character.
//     For instance, "ckpr5gut7qls715olr70_remap.uf2" -> "ckpr5gut7qls715olr70_remap_1580000000.uf2"
//  3. Return the firmware file name with the timestamp.
func CreateFirmwareFileNameWithTimestamp(firmwareFileName string) string {
	if firmwareFileName == "" {
		return ""
	}
	epoch := strconv.FormatInt(time.Now().Unix(), 10)
	return firmwareFileName[:len(firmwareFileName)-len(filepath.Ext(firmwareFileName))] + "_" + epoch + filepath.Ext(firmwareFileName)
}

func CreateFirmwareFilePath(qmkFirmwareBaseDirectoryPath string, qmkFirmwareVersion string, firmwareFileName string) string {
	return filepath.Join(qmkFirmwareBaseDirectoryPath+qmkFirmwareVersion, CreateFirmwareFileNameWithTimestamp(firmwareFileName))
}
