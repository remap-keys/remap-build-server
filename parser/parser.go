package parser

import "regexp"

func ReplaceParameters(input string, replacements map[string]string) string {
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
