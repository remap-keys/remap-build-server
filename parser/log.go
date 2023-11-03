package parser

import (
	"errors"
	"regexp"
)

// FetchFirmwareFileName fetches the firmware file name from the stdout of the `qmk compile` command.
// The stdout string includes like the followings:
// * "Copying ckpr5gut7qls715olr70_remap.uf2 to qmk_firmware folder"
// * "Copying ckpr5gut7qls715olr70_remap.hex to qmk_firmware folder"
// * "Copying ckpr5gut7qls715olr70_remap.bin to qmk_firmware folder"
// That is, the regular expression of the firmware file is /[a-zA-Z0-9_]+\.[a-zA-Z0-9]+/
// The file extension of the firmware file name is any string after the "." character excluding spaces.
// If these patterns are not matched, it returns an error.
func FetchFirmwareFileName(stdout string) (string, error) {
	re := regexp.MustCompile(`Copying ([a-zA-Z0-9_]+\.[a-zA-Z0-9]+) to qmk_firmware folder`)
	matches := re.FindStringSubmatch(stdout)
	if len(matches) != 2 {
		return "", errors.New("Failed to fetch the firmware file name from the stdout of the `qmk compile` command.")
	}
	return matches[1], nil
}
