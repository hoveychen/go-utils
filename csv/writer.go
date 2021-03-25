// Package csv provides CsvReader and CsvWriter to process csv format file
// in the struct declaration style.
package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"

	goutils "github.com/hoveychen/go-utils"
)

const defaultSliceDelimiter = "\n"
const defaultTagDelimiter = ","

var bomUtf8 = []byte{0xEF, 0xBB, 0xBF}

// CsvWriter extends the encoding/csv writer, supporting writting struct, and
// shortcut to write to a file.
type CsvWriter struct {
	sync.Mutex
	*csv.Writer
	Headers        []string
	file           *os.File
	fieldIdx       []string
	sliceDelimiter string
	skipJsonNull   bool
}

func NewCsvWriter(w io.Writer) *CsvWriter {
	w.Write(bomUtf8)
	return &CsvWriter{
		Writer:         csv.NewWriter(w),
		sliceDelimiter: defaultSliceDelimiter,
		skipJsonNull:   true,
	}
}

func NewFileCsvWriter(filename string) *CsvWriter {
	file, err := os.Create(filename)
	if err != nil {
		goutils.LogError(err)
		return nil
	}
	file.Write(bomUtf8)
	return &CsvWriter{
		Writer:         csv.NewWriter(file),
		file:           file,
		sliceDelimiter: defaultSliceDelimiter,
		skipJsonNull:   true,
	}
}

func (w *CsvWriter) buildFieldIndex(val reflect.Value) {
	w.fieldIdx = []string{}
	for i := 0; i < val.Type().NumField(); i++ {
		field := val.Type().Field(i)
		if field.PkgPath != "" {
			// Unexported field will have PkgPath.
			continue
		}
		if w.skipJsonNull && field.Tag.Get("json") == "-" {
			continue
		}
		tag := field.Tag.Get("csv")
		var name string
		if tag == "" {
			name = field.Name
		} else if tag == "-" {
			continue
		} else {
			name = tag
		}

		w.Headers = append(w.Headers, name)
		w.fieldIdx = append(w.fieldIdx, field.Name)
	}
}

func (w *CsvWriter) SetSliceDelimiter(delim string) {
	w.sliceDelimiter = delim
}

func (w *CsvWriter) SetSkipJsonNull(skip bool) {
	w.skipJsonNull = skip
}

func (w *CsvWriter) WriteStruct(i interface{}) error {
	w.Lock()
	defer w.Unlock()
	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.New("Input need to be a struct")
	}

	if w.Headers == nil {
		w.buildFieldIndex(val)
		w.Write(w.Headers)
	}

	out := []string{}
	for _, name := range w.fieldIdx {
		v := val.FieldByName(name)
		switch v.Kind() {
		case reflect.Slice:
			var segs []string
			for i := 0; i < v.Len(); i++ {
				segs = append(segs, fmt.Sprint(v.Index(i).Interface()))
			}
			out = append(out, strings.Join(segs, w.sliceDelimiter))
		default:
			out = append(out, fmt.Sprint(v.Interface()))
		}
	}
	w.Write(out)
	return nil
}

func (w *CsvWriter) Close() error {
	w.Lock()
	defer w.Unlock()
	w.Flush()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
