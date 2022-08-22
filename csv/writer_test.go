package csv

import (
	"bytes"
	"fmt"
	"testing"
)

func ExampleCsvWriter() {
	type TestStruct struct {
		Int        int
		String     string
		TagInt     int `csv:"tag_i"`
		anonyInt   int
		HiddenInt  int    `csv:"-"`
		NextString string `csv:"nextstring"`
	}
	t := TestStruct{
		1, "2", 3, 4, 5, "6",
	}

	buf := &bytes.Buffer{}
	w := NewCsvWriter(buf, WithAppendBOM(false))
	w.WriteStruct(t)
	w.Flush()
	fmt.Println(buf.String())
	// Output:
	// Int,String,tag_i,nextstring
	// 1,2,3,6
}

func TestCsvWriter_WriteStruct(t *testing.T) {
	var buf bytes.Buffer

	w := NewCsvWriter(&buf, WithAppendBOM(false), WithSliceDelimiter(","))

	type MultipleSpan struct {
		Names   []string `csv:"name,span=3"`
		Departs []string `csv:"departs,limit=5"`
		Age     int      `csv:"age"`
		Absent  string   `json:"-"`
	}

	w.WriteStruct(&MultipleSpan{
		Names:   []string{"a", "b", "c", "d"},
		Age:     18,
		Departs: []string{"111", "222"},
		Absent:  "Should not display",
	})

	w.WriteStruct(&MultipleSpan{
		Names:  []string{"e", "f"},
		Age:    20,
		Absent: "Should not display",
	})

	w.Close()

	expected := `name,name,name,departs,age
a,b,c,"111,2",18
e,f,,,20
`
	if buf.String() != expected {
		t.Errorf("Expect:\n%s\nActual:\n%s", expected, buf.String())
	}

}
