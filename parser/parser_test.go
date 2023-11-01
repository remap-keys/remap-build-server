package parser

import "testing"

func Test_ReplaceParameters_EmptySource(t *testing.T) {
	actual := ReplaceParameters("", map[string]string{})
	expected := ""
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParameters_NoParameters(t *testing.T) {
	actual := ReplaceParameters("foo", map[string]string{})
	expected := "foo"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParameters_OneParameter(t *testing.T) {
	actual := ReplaceParameters(`<remap name="foo" type="select" />`, map[string]string{"foo": "bar"})
	expected := "bar"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParameters_TwoParameters(t *testing.T) {
	actual := ReplaceParameters(`<remap name="foo" type="select" /><remap name="bar" type="number" />`, map[string]string{"foo": "bar", "bar": "baz"})
	expected := "barbaz"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParameters_OneParameterWithOption(t *testing.T) {
	actual := ReplaceParameters(`<remap name="foo" type="select" options="foo,bar,baz" />`, map[string]string{"foo": "bar"})
	expected := "bar"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParameters_OneParameterAndOneNonParameter(t *testing.T) {
	actual := ReplaceParameters(`<remap name="foo" type="select" />bar`, map[string]string{"foo": "baz"})
	expected := "bazbar"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParameters_OneParameterAndOneNonParameterAndOneParameter(t *testing.T) {
	actual := ReplaceParameters(`<remap name="foo" type="select" />bar<remap name="bar" type="number" />`, map[string]string{"foo": "baz", "bar": "qux"})
	expected := "bazbarqux"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParameters_OneParameterAndOneMissingParameter(t *testing.T) {
	actual := ReplaceParameters(`<remap name="foo" type="select" />`, map[string]string{"bar": "baz"})
	expected := ""
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParameters_OneParameterAndOneMissingParameterAndOneParameter(t *testing.T) {
	actual := ReplaceParameters(`<remap name="foo" type="select" /><remap name="bar" type="number" />`, map[string]string{"foo": "baz"})
	expected := "baz"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParameters_OneParameterAndOneMissingParameterAndOneNonParameter(t *testing.T) {
	actual := ReplaceParameters(`<remap name="foo" type="select" />bar`, map[string]string{"bar": "baz"})
	expected := "bar"
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ReplaceParameters_TwoParametersAndMultipleLinesSource(t *testing.T) {
	actual := ReplaceParameters(`<remap name="foo" type="select" />
<remap name="bar" type="number" />`, map[string]string{"foo": "baz", "bar": "qux"})
	expected := `baz
qux`
	if actual != expected {
		t.Error("Expected", expected, "but got", actual)
	}
}

func Test_ExtractParameters_EmptySource(t *testing.T) {
	actual := ExtractParameters("")
	var expected []Parameter
	if len(actual) != len(expected) {
		t.Error("Expected 0 parameters")
	}
}

func Test_ExtractParameters_NoParameters(t *testing.T) {
	actual := ExtractParameters("foo")
	var expected []Parameter
	if len(actual) != len(expected) {
		t.Error("Expected 0 parameters")
	}
}

func Test_ExtractParameters_OneParameter(t *testing.T) {
	actual := ExtractParameters(`<remap name="foo" type="select" />`)
	expected := []Parameter{
		{"foo", "select", []string{}, 0, 34},
	}
	if len(actual) != len(expected) {
		t.Error("Expected 1 parameter")
	}
	if actual[0].Name != expected[0].Name {
		t.Error("Expected", expected[0].Name, "but got", actual[0].Name)
	}
	if actual[0].Type != expected[0].Type {
		t.Error("Expected", expected[0].Type, "but got", actual[0].Type)
	}
	if actual[0].StartPosition != expected[0].StartPosition {
		t.Error("Expected", expected[0].StartPosition, "but got", actual[0].StartPosition)
	}
	if actual[0].EndPosition != expected[0].EndPosition {
		t.Error("Expected", expected[0].EndPosition, "but got", actual[0].EndPosition)
	}
}

func Test_ExtractParameters_TwoParameters(t *testing.T) {
	actual := ExtractParameters(`<remap name="foo" type="select" /><remap name="bar" type="number" />`)
	expected := []Parameter{
		{"foo", "select", []string{}, 0, 34},
		{"bar", "number", []string{}, 34, 68},
	}
	if len(actual) != len(expected) {
		t.Error("Expected 2 parameters")
	}
	if actual[0].Name != expected[0].Name {
		t.Error("Expected", expected[0].Name, "but got", actual[0].Name)
	}
	if actual[0].Type != expected[0].Type {
		t.Error("Expected", expected[0].Type, "but got", actual[0].Type)
	}
	if actual[0].StartPosition != expected[0].StartPosition {
		t.Error("Expected", expected[0].StartPosition, "but got", actual[0].StartPosition)
	}
	if actual[0].EndPosition != expected[0].EndPosition {
		t.Error("Expected", expected[0].EndPosition, "but got", actual[0].EndPosition)
	}
	if actual[1].Name != expected[1].Name {
		t.Error("Expected", expected[1].Name, "but got", actual[1].Name)
	}
	if actual[1].Type != expected[1].Type {
		t.Error("Expected", expected[1].Type, "but got", actual[1].Type)
	}
	if actual[1].StartPosition != expected[1].StartPosition {
		t.Error("Expected", expected[1].StartPosition, "but got", actual[1].StartPosition)
	}
	if actual[1].EndPosition != expected[1].EndPosition {
		t.Error("Expected", expected[1].EndPosition, "but got", actual[1].EndPosition)
	}
}
