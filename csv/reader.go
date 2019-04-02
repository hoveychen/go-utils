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

	multierror "github.com/hashicorp/go-multierror"
	goutils "github.com/hoveychen/go-utils"
)

type CsvReader struct {
	*csv.Reader
	Headers        []string
	fieldIdx       []string
	file           *os.File
	sliceDelimiter string
	tagDelimiter   string
}

func NewCsvReader(r io.Reader) *CsvReader {
	return &CsvReader{
		Reader:         csv.NewReader(r),
		sliceDelimiter: defaultSliceDelimiter,
		tagDelimiter:   defaultTagDelimiter,
	}
}

func NewFileCsvReader(filename string) *CsvReader {
	file, err := os.Open(filename)
	if err != nil {
		goutils.LogError(err)
		return nil
	}
	return &CsvReader{
		Reader:         csv.NewReader(file),
		file:           file,
		sliceDelimiter: defaultSliceDelimiter,
	}
}

func (r *CsvReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

func (r *CsvReader) SetSliceDelimiter(delim string) {
	r.sliceDelimiter = delim
}

func (r *CsvReader) SetTagDelimiter(delim string) {
	r.tagDelimiter = delim
}

func (r *CsvReader) buildFieldIndex(val reflect.Value, row []string) {
	colDict := map[string]string{}
	for i := 0; i < val.Type().NumField(); i++ {
		field := val.Type().Field(i)
		if field.PkgPath != "" {
			// Unexported field will have PkgPath.
			continue
		}
		textTags := field.Tag.Get("csv")
		var tags []string

		if textTags == "" {
			tags = []string{field.Name}
		} else if textTags == "-" {
			continue
		} else {
			tags = strings.Split(textTags, r.tagDelimiter)
		}

		for _, name := range tags {
			name = strings.TrimSpace(strings.ToLower(name))
			if colDict[name] != "" {
				goutils.LogError("Duplicated field name", name)
				continue
			} else {
				colDict[name] = field.Name
			}
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

	var allError error
	for idx, col := range r.fieldIdx {
		if idx >= len(row) {
			// Should never be here.
			continue
		}
		if col == "" {
			continue
		}
		if row[idx] == "" {
			continue
		}
		v := val.FieldByName(col)
		switch v.Kind() {
		case reflect.String:
			// Try to parse the string to correspond type.
			// NOTE: Using fmt.Sscanf("%v") will only parse the first
			// space-delimited token. If the cell contains like "123 abc",
			// only "123" will be parsed, while "abc" ignored.
			v.SetString(row[idx])
		case reflect.Slice:
			segs := strings.Split(row[idx], r.sliceDelimiter)
			slice := reflect.MakeSlice(v.Type(), len(segs), len(segs))
			v.Set(slice)
			for idx, s := range segs {
				switch v.Type().Elem().Kind() {
				case reflect.String:
					slice.Index(idx).SetString(s)
				default:
					_, err := fmt.Sscanf(s, "%v", slice.Index(idx).Addr().Interface())
					if err != nil && err != io.EOF {
						allError = multierror.Append(allError, err)
					}
				}
			}
		case reflect.Map:
			if v.IsNil() {
				// Create map first.
				m := reflect.MakeMap(v.Type())
				v.Set(m)
			}
			columnName := r.Headers[idx]
			value := reflect.New(v.Type().Elem())
			_, err := fmt.Sscanf(row[idx], "%v", value.Interface())
			if err != nil && err != io.EOF {
				allError = multierror.Append(allError, err)
				continue
			}
			v.SetMapIndex(reflect.ValueOf(columnName), value.Elem())
		default:
			_, err := fmt.Sscanf(row[idx], "%v", v.Addr().Interface())
			if err != nil && err != io.EOF {
				allError = multierror.Append(allError, err)
			}
		}
	}
	return allError
}
