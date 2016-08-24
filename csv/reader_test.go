package csv

import (
	"bytes"
	"fmt"
)

func ExampleCsvReader() {
	type TestStruct struct {
		Int        int
		String     string
		TagInt     int `csv:"tag_i"`
		anonyInt   int
		HiddenInt  int    `csv:"-"`
		NextString string `csv:"nextstring"`
	}
	input := `Int,String,tag_i,nextstring
1,2,3,6`
	t := &TestStruct{}

	buf := bytes.NewBufferString(input)
	r := NewCsvReader(buf)
	r.ReadStruct(t)
	fmt.Printf("%+v", t)
	// Output:
	// &{Int:1 String:2 TagInt:3 anonyInt:0 HiddenInt:0 NextString:6}
}
