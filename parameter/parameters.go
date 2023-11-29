package parameter

import (
	"encoding/json"
	"regexp"
	"remap-keys.app/remap-build-server/common"
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
func ReplaceParameters(files []*common.FirmwareFile, parameterFileMap map[string]map[string]string) []*common.FirmwareFile {
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

// ParseParameterJson parses the ParameterJson string.
func ParseParameterJson(parametersJson string) (*common.ParametersJson, error) {
	var parseResult map[string]interface{}
	err := json.Unmarshal([]byte(parametersJson), &parseResult)
	if err != nil {
		return nil, err
	}
	versionValue := parseResult["version"]
	var version int8
	if versionValue == nil {
		version = 1
	} else {
		version = int8(versionValue.(float64))
	}
	if version == 2 {
		var result common.ParametersJson
		err = json.Unmarshal([]byte(parametersJson), &result)
		if err != nil {
			return nil, err
		}
		return &result, nil
	}
	// Convert the version 1 to the version 2.
	var parseResultByVersion1 common.ParametersJsonVersion1
	err = json.Unmarshal([]byte(parametersJson), &parseResultByVersion1)
	if err != nil {
		return nil, err
	}
	var result common.ParametersJson
	result.Version = 1
	keyboardMap := map[string]common.ParameterValue{}
	for key, value := range parseResultByVersion1.Keyboard {
		keyboardMap[key] = common.ParameterValue{
			Type:       "parameters",
			Parameters: value,
			Code:       "",
		}
	}
	result.Keyboard = keyboardMap
	keymapMap := map[string]common.ParameterValue{}
	for key, value := range parseResultByVersion1.Keymap {
		keymapMap[key] = common.ParameterValue{
			Type:       "parameters",
			Parameters: value,
			Code:       "",
		}
	}
	result.Keymap = keymapMap
	return &result, nil
}
