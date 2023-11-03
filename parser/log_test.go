package parser

import "testing"

func Test_FetchFirmwareFileName_EmptySource(t *testing.T) {
	_, err := FetchFirmwareFileName("")
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func Test_FetchFirmwareFileName_NoFirmwareFileName(t *testing.T) {
	_, err := FetchFirmwareFileName("foo")
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func Test_FetchFirmwareFileName_FirmwareFileNameWithUF2Extension(t *testing.T) {
	actual, err := FetchFirmwareFileName("Copying ckpr5gut7qls715olr70_remap.uf2 to qmk_firmware folder")
	if err != nil {
		t.Error("Expected nil but got", err)
	}
	expected := "ckpr5gut7qls715olr70_remap.uf2"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_FetchFirmwareFileName_FirmwareFileNameWithHexExtension(t *testing.T) {
	actual, err := FetchFirmwareFileName("Copying ckpr5gut7qls715olr70_remap.hex to qmk_firmware folder")
	if err != nil {
		t.Error("Expected nil but got", err)
	}
	expected := "ckpr5gut7qls715olr70_remap.hex"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_FetchFirmwareFileName_FirmwareFileNameInMultipleLineString(t *testing.T) {
	actual, err := FetchFirmwareFileName("foo\nCopying ckpr5gut7qls715olr70_remap.hex to qmk_firmware folder\nbar")
	if err != nil {
		t.Error("Expected nil but got", err)
	}
	expected := "ckpr5gut7qls715olr70_remap.hex"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}
