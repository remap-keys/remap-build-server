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
