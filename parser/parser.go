package parser

import "regexp"

type Parameter struct {
	Name          string
	Type          string
	Options       []string
	StartPosition int
	EndPosition   int
}

func ExtractParameters(source string) []Parameter {
	var parameters []Parameter
	regex := regexp.MustCompile(`<remap name="(?P<name>[^"]+)" type="(?P<type>[^"]+)"(?: options="(?P<options>[^"]+)")? />`)
	matches := regex.FindAllStringSubmatchIndex(source, -1)
	for _, match := range matches {
		name := source[match[2]:match[3]]
		typ := source[match[4]:match[5]]
		var options []string
		if match[6] != -1 && match[7] != -1 {
			options = splitOptions(source[match[6]:match[7]])
		}
		parameters = append(parameters, Parameter{
			Name:          name,
			Type:          typ,
			Options:       options,
			StartPosition: match[0],
			EndPosition:   match[1],
		})
	}
	return parameters
}

func splitOptions(options string) []string {
	return regexp.MustCompile(`[^,]+`).FindAllString(options, -1)
}

func ReplaceParameters(source string, values map[string]string) string {
	regex := regexp.MustCompile(`<remap name="([^"]+)" type="([^"]+)"(?: options="([^"]+)")? />`)
	return regex.ReplaceAllStringFunc(source, func(match string) string {
		matches := regex.FindStringSubmatch(match)
		name := matches[1]
		if value, exists := values[name]; exists {
			return value
		}
		return ""
	})
}
