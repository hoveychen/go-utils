package goutils

import "testing"

func TestSprintt(t *testing.T) {
	tmpl := "Hello. My name is {{.Name}}, you can call me {{.Nickname}}"

	data := map[string]string{
		"Name":     "Harry",
		"Nickname": "H",
	}

	expected := "Hello. My name is Harry, you can call me H"
	actual := Sprintt(tmpl, data)
	if actual != expected {
		t.Error("Actual", actual, "\nExpected", expected)
	}

	// Tests if tmpl cache mass up.
	actual = Sprintt(tmpl, data)
	if actual != expected {
		t.Error("Actual", actual, "\nExpected", expected)
	}

	tmpl2 := "{{.Nickname}} is {{.Name}}"
	expected2 := "H is Harry"
	actual2 := Sprintt(tmpl2, data)
	if actual2 != expected2 {
		t.Error("Actual", actual2, "\nExpected", expected2)
	}
}
