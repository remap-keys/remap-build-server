package parameter

import "testing"

func Test_ReplaceParametersInString_EmptySource(t *testing.T) {
	actual := ReplaceParametersInString("", map[string]string{})
	expected := ""
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParametersInString_NoParameters(t *testing.T) {
	actual := ReplaceParametersInString("foo", map[string]string{})
	expected := "foo"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParametersInString_OneParameter(t *testing.T) {
	actual := ReplaceParametersInString(`<remap name="foo" type="select" />`, map[string]string{"foo": "bar"})
	expected := "bar"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParametersInString_TwoParameters(t *testing.T) {
	actual := ReplaceParametersInString(`<remap name="foo" type="select" /><remap name="bar" type="number" />`, map[string]string{"foo": "bar", "bar": "baz"})
	expected := "barbaz"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParametersInString_OneParameterWithOption(t *testing.T) {
	actual := ReplaceParametersInString(`<remap name="foo" type="select" options="foo,bar,baz" />`, map[string]string{"foo": "bar"})
	expected := "bar"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParametersInString_OneParameterAndOneNonParameter(t *testing.T) {
	actual := ReplaceParametersInString(`<remap name="foo" type="select" />bar`, map[string]string{"foo": "baz"})
	expected := "bazbar"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParametersInString_OneParameterAndOneNonParameterAndOneParameter(t *testing.T) {
	actual := ReplaceParametersInString(`<remap name="foo" type="select" />bar<remap name="bar" type="number" />`, map[string]string{"foo": "baz", "bar": "qux"})
	expected := "bazbarqux"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParametersInString_OneParameterAndOneMissingParameter(t *testing.T) {
	actual := ReplaceParametersInString(`<remap name="foo" type="select" />`, map[string]string{"bar": "baz"})
	expected := ""
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParametersInString_OneParameterAndOneMissingParameterAndOneParameter(t *testing.T) {
	actual := ReplaceParametersInString(`<remap name="foo" type="select" /><remap name="bar" type="number" />`, map[string]string{"foo": "baz"})
	expected := "baz"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParametersInString_OneParameterAndOneMissingParameterAndOneNonParameter(t *testing.T) {
	actual := ReplaceParametersInString(`<remap name="foo" type="select" />bar`, map[string]string{"bar": "baz"})
	expected := "bar"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParametersInString_TwoParametersAndMultipleLinesSource(t *testing.T) {
	actual := ReplaceParametersInString(`<remap name="foo" type="select" />
<remap name="bar" type="number" />`, map[string]string{"foo": "baz", "bar": "qux"})
	expected := `baz
qux`
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ParseParameterJson_EmptySource(t *testing.T) {
	_, err := ParseParameterJson("")
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func Test_ParseParameterJson_InvalidJson(t *testing.T) {
	_, err := ParseParameterJson("foo")
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func Test_ParseParameterJson_Version1(t *testing.T) {
	json := `{
				"keyboard": {
					"file1": {
						"name1": "value1"
					}
				},
				"keymap": {
					"file2": {
						"name2": "value2"
					}
				}
			}`
	actual, err := ParseParameterJson(json)
	if err != nil {
		t.Error("Expected nil but got", err)
	}
	if actual.Version != 1 {
		t.Error("Expected 1 but got", actual.Version)
	}
	if actual.Keyboard["file1"].Type != "parameters" {
		t.Error("Expected parameters but got", actual.Keyboard["file1"].Type)
	}
	if actual.Keyboard["file1"].Parameters["name1"] != "value1" {
		t.Error("Expected value1 but got", actual.Keyboard["file1"].Parameters["name1"])
	}
	if actual.Keyboard["file1"].Code != "" {
		t.Error("Expected empty string but got", actual.Keyboard["file1"].Code)
	}
	if actual.Keymap["file2"].Type != "parameters" {
		t.Error("Expected parameters but got", actual.Keymap["file2"].Type)
	}
	if actual.Keymap["file2"].Parameters["name2"] != "value2" {
		t.Error("Expected value2 but got", actual.Keymap["file2"].Parameters["name2"])
	}
	if actual.Keymap["file2"].Code != "" {
		t.Error("Expected empty string but got", actual.Keymap["file2"].Code)
	}
}

func Test_ParseParameterJson_Version2(t *testing.T) {
	json := `{
				"version": 2,
				"keyboard": {
					"file1": {
						"type": "parameters",
						"parameters": {
							"name1": "value1"
						}
					}
				},
				"keymap": {
					"file2": {
						"type": "code",
						"code": "code2"
					}
				}
			}`
	actual, err := ParseParameterJson(json)
	if err != nil {
		t.Error("Expected nil but got", err)
	}
	if actual.Version != 2 {
		t.Error("Expected 2 but got", actual.Version)
	}
	if actual.Keyboard["file1"].Type != "parameters" {
		t.Error("Expected parameters but got", actual.Keyboard["file1"].Type)
	}
	if actual.Keyboard["file1"].Parameters["name1"] != "value1" {
		t.Error("Expected value1 but got", actual.Keyboard["file1"].Parameters["name1"])
	}
	if actual.Keyboard["file1"].Code != "" {
		t.Error("Expected empty string but got", actual.Keyboard["file1"].Code)
	}
	if actual.Keymap["file2"].Type != "code" {
		t.Error("Expected code but got", actual.Keymap["file2"].Type)
	}
	if len(actual.Keyboard["file2"].Parameters) != 0 {
		t.Error("Expected 0 but got", len(actual.Keyboard["file2"].Parameters))
	}
	if actual.Keymap["file2"].Code != "code2" {
		t.Error("Expected code2 but got", actual.Keymap["file2"].Code)
	}
}
