package build

import (
	"regexp"
	"testing"
)

func Test_CreateFirmwareFileNameWithTimestamp_WithoutExt(t *testing.T) {
	actual := CreateFirmwareFileNameWithTimestamp("foo")
	if actual == "" {
		t.Error("Expected not empty string but got", actual)
	}
	if !validateFileNameWithTimestampFormat(actual) {
		t.Error("Expected valid file name but got", actual)
	}
}

func Test_CreateFirmwareFileNameWithTimestamp_WithExt(t *testing.T) {
	actual := CreateFirmwareFileNameWithTimestamp("foo.hex")
	if actual == "" {
		t.Error("Expected not empty string but got", actual)
	}
	if !validateFileNameWithTimestampFormat(actual) {
		t.Error("Expected valid file name but got", actual)
	}
}

func Test_CreateFirmwareFileNameWithTimestamp_Empty(t *testing.T) {
	actual := CreateFirmwareFileNameWithTimestamp("")
	if actual != "" {
		t.Error("Expected empty string but got", actual)
	}
}

func validateFileNameWithTimestampFormat(source string) bool {
	pattern := `[a-zA-Z0-9_]+_[0-9]+(.[a-zA-Z0-9]+)?`
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}
	return re.MatchString(source)
}
