package csv

import (
	"bytes"
	"fmt"
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
	w := NewCsvWriter(buf)
	w.WriteStruct(t)
	w.Flush()
	fmt.Println(buf.String())
	// Output:
	// Int,String,tag_i,nextstring
	// 1,2,3,6
}
