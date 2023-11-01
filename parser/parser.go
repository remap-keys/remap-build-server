package parser

import "regexp"

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
