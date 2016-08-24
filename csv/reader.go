// Package csv provides CsvReader and CsvWriter to process csv format file
// in the struct declaration style.
package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/hoveychen/go-utils"
	"io"
	"os"
	"reflect"
	"strings"
)

type CsvReader struct {
	*csv.Reader
	Headers  []string
	fieldIdx []string
	file     *os.File
}

func NewCsvReader(r io.Reader) *CsvReader {
	return &CsvReader{
		Reader: csv.NewReader(r),
	}
}

func NewFileCsvReader(filename string) *CsvReader {
	file, err := os.Open(filename)
	if err != nil {
		goutils.LogError(err)
		return nil
	}
	return &CsvReader{
		Reader: csv.NewReader(file),
		file:   file,
	}
}

func (r *CsvReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

func (r *CsvReader) buildFieldIndex(val reflect.Value, row []string) {
	colDict := map[string]string{}
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
		name = strings.TrimSpace(strings.ToLower(name))

		_, exists := colDict[name]
		if exists {
			goutils.LogError("Duplicated field name", name)
			continue
		} else {
			colDict[name] = field.Name
		}
	}

	r.Headers = row
	for _, h := range row {
		col := strings.TrimSpace(strings.ToLower(h))
		name, exists := colDict[col]
		if !exists {
			r.fieldIdx = append(r.fieldIdx, "")
		} else {
			r.fieldIdx = append(r.fieldIdx, name)
		}
	}
}

func (r *CsvReader) ReadStruct(i interface{}) error {
	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.New("Input need to be a struct")
	}

	row, err := r.Read()
	if err != nil {
		return err
	}

	if r.Headers == nil {
		r.buildFieldIndex(val, row)
		row, err = r.Read()
		if err != nil {
			return err
		}
	}

	for idx, col := range r.fieldIdx {
		if idx >= len(row) {
			// Should never be here.
			continue
		}
		if col == "" {
			continue
		}
		v := val.FieldByName(col)
		if v.Kind() == reflect.String {
			v.SetString(row[idx])
		} else {
			// Try to parse the string to correspond type.
			// NOTE: Using fmt.Sscanf("%v") will only parse the first
			// space-delimited token. If the cell contains like "123 abc",
			// only "123" will be parsed, while "abc" ignored.
			_, err := fmt.Sscanf(row[idx], "%v", v.Addr().Interface())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
