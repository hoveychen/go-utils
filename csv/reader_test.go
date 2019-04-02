package csv

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	goutils "github.com/hoveychen/go-utils"
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

func compIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func compStrSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSliceStruct(t *testing.T) {
	type TestStruct struct {
		Int         int      `csv:"int"`
		String      string   `csv:"string"`
		StringSlice []string `csv:"str_slice"`
		IntSlice    []int    `csv:"int_slice"`
	}

	input := `int,string,str_slice,int_slice
1,hovey,"foo:bar:","2:3:5:8:13"
2,chen,,"1024:"`

	expected := []*TestStruct{
		{1, "hovey", []string{"foo", "bar", ""}, []int{2, 3, 5, 8, 13}},
		{2, "chen", []string{}, []int{1024, 0}},
	}

	r := NewCsvReader(bytes.NewBufferString(input))
	defer r.Close()
	r.SetSliceDelimiter(":")

	for i := 0; i < len(expected); i++ {
		st := &TestStruct{}
		if err := r.ReadStruct(st); err != nil {
			t.Errorf("ReadStruct Line:%d, err=%v", i, err)
			continue
		}

		exp := expected[i]
		if exp.Int != st.Int {
			t.Errorf("Int not same: expected %d actual %d", exp.Int, st.Int)
		}
		if exp.String != st.String {
			t.Errorf("String not same: expected %s actual %s", exp.String, st.String)
		}
		if !compIntSlice(exp.IntSlice, st.IntSlice) {
			t.Errorf("IntSlice not same: expected %v actual %v", exp.IntSlice, st.IntSlice)
		}
		if !compStrSlice(exp.StringSlice, st.StringSlice) {
			t.Errorf("StringSlice not same: expected %v actual %v", exp.StringSlice, st.StringSlice)
		}
	}
	st := &TestStruct{}
	if err := r.ReadStruct(st); err != io.EOF {
		t.Error("Not correctly output EOF")
	}
}

func TestMapStruct(t *testing.T) {
	type TestStruct struct {
		Item     string             `csv:"item"`
		Currency map[string]string  `csv:"US,CN,FR"`
		Price    map[string]float64 `csv:"US-price,CN-price,FR-price"`
	}

	input := `US,US-price,FR,FR-price,GB,GB-price,item
USD,49.99,EUR,59.99,GBP,39.99,small set
USD,89.99,EUR,99.99,GBP,79.99,large set`

	expected := []*TestStruct{
		{"small set", map[string]string{"US": "USD", "FR": "EUR"}, map[string]float64{"US-price": 49.99, "FR-price": 59.99}},
		{"large set", map[string]string{"US": "USD", "FR": "EUR"}, map[string]float64{"US-price": 89.99, "FR-price": 99.99}},
	}

	r := NewCsvReader(bytes.NewBufferString(input))
	defer r.Close()
	r.SetTagDelimiter(",")

	for i := 0; i < len(expected); i++ {
		st := &TestStruct{}
		if err := r.ReadStruct(st); err != nil {
			t.Errorf("ReadStruct Line:%d, err=%v", i, err)
			continue
		}

		exp := expected[i]
		if goutils.Jsonify(st) != goutils.Jsonify(exp) {
			t.Error("Expect:\n", goutils.Jsonify(exp), "\nActual:\n", goutils.Jsonify(st))
		}
	}
	st := &TestStruct{}
	if err := r.ReadStruct(st); err != io.EOF {
		t.Error("Not correctly output EOF")
	}
}
