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
	"sync"

	"github.com/hoveychen/go-utils"
)

// CsvWriter extends the encoding/csv writer, supporting writting struct, and
// shortcut to write to a file.
type CsvWriter struct {
	sync.Mutex
	*csv.Writer
	Headers  []string
	file     *os.File
	fieldIdx []string
}

func NewCsvWriter(w io.Writer) *CsvWriter {
	return &CsvWriter{
		Writer: csv.NewWriter(w),
	}
}

func NewFileCsvWriter(filename string) *CsvWriter {
	file, err := os.Create(filename)
	if err != nil {
		goutils.LogError(err)
		return nil
	}
	return &CsvWriter{
		Writer: csv.NewWriter(file),
		file:   file,
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
		v := val.FieldByName(name).Interface()
		out = append(out, fmt.Sprintf("%v", v))
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
