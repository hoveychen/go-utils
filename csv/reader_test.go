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

func TestMapStruct2(t *testing.T) {
	type Caterow struct {
		CateID    string         `csv:"cate_id"`
		Weight    float64        `csv:"weight kg"`
		Sellables map[string]int `csv:"INTL,US,AE,SA,IN,ID,TH,VN,MY,SG,PH,AT,AU,BE,CA,CH,CN,CZ,DE,DK,ES,FI,FR,GB,HK,IE,IL,IT,JP,KR,KW,MO,MX,NL,NO,NZ,PL,PT,QA,RU,SE,TW,TR,UA,ZA"`
	}

	input := `cate_id,cn,SKU数量-20190111,en,parent_id,weight kg,leaf,lv1,lv2,lv3,lv4,lv5,platform,INTL,US,AE,SA,IN,ID,TH,VN,MY,SG,PH,AT,AU,BE,CA,CH,CN,CZ,DE,DK,ES,FI,FR,GB,HK,IE,IL,IT,JP,KR,KW,MO,MX,NL,NO,NZ,PL,PT,QA,RU,SE,TW,TR,UA,ZA
tb:16881031910,连衣裙,161542,tb:16881031910,tb:168810166,0.25,TRUE,女装,连衣裙,,,,1688,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1`

	expected := []*Caterow{
		{
			"tb:16881031910", 0.25, map[string]int{
				"INTL": 1, "US": 1, "AE": 1, "SA": 1, "IN": 1, "ID": 1, "TH": 1, "VN": 1, "MY": 1, "SG": 1, "PH": 1, "AT": 1, "AU": 1, "BE": 1, "CA": 1, "CH": 1,
				"CN": 0, "CZ": 1, "DE": 1, "DK": 1, "ES": 1, "FI": 1, "FR": 1, "GB": 1, "HK": 1, "IE": 1, "IL": 1, "IT": 1, "JP": 1, "KR": 1, "KW": 1, "MO": 1,
				"MX": 1, "NL": 1, "NO": 1, "NZ": 1, "PL": 1, "PT": 1, "QA": 1, "RU": 1, "SE": 1, "TW": 1, "TR": 1, "UA": 1, "ZA": 1,
			},
		},
	}

	r := NewCsvReader(bytes.NewBufferString(input))
	defer r.Close()
	r.SetTagDelimiter(",")

	for i := 0; i < len(expected); i++ {
		st := &Caterow{}
		if err := r.ReadStruct(st); err != nil {
			t.Errorf("ReadStruct Line:%d, err=%v", i, err)
			continue
		}

		exp := expected[i]
		if goutils.Jsonify(st) != goutils.Jsonify(exp) {
			t.Error("Expect:\n", goutils.Jsonify(exp), "\nActual:\n", goutils.Jsonify(st))
		}
	}
	st := &Caterow{}
	if err := r.ReadStruct(st); err != io.EOF {
		t.Error("Not correctly output EOF")
	}
}
