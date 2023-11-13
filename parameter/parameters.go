package parameter

import (
	"regexp"
	"remap-keys.app/remap-build-server/database"
)

func ReplaceParametersInString(input string, replacements map[string]string) string {
	re := regexp.MustCompile(`<remap\s+([^/>]+)\/>`)
	return re.ReplaceAllStringFunc(input, func(m string) string {
		attrMatcher := regexp.MustCompile(`(\w+)="([^"]+)"`)
		attrs := attrMatcher.FindAllStringSubmatch(m, -1)
		attrMap := make(map[string]string)
		for _, pair := range attrs {
			attrMap[pair[1]] = pair[2]
		}

		if val, ok := replacements[attrMap["name"]]; ok {
			return val
		}
		return ""
	})
}

// ReplaceParameters replaces the parameters in the keyboard files.
func ReplaceParameters(files []*database.FirmwareFile, parameterFileMap map[string]map[string]string) []*database.FirmwareFile {
	for _, file := range files {
		parameterMap := parameterFileMap[file.ID]
		if parameterMap == nil {
			// If there is no parameter map for the firmware file, skip this file.
			continue
		}
		newContent := ReplaceParametersInString(file.Content, parameterMap)
		file.Content = newContent
	}
	return files
}
