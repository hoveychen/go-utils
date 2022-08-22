// Package csv provides CsvReader and CsvWriter to process csv format file
// in the struct declaration style.
package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"unicode"

	multierror "github.com/hashicorp/go-multierror"
	goutils "github.com/hoveychen/go-utils"
	"github.com/pkg/errors"
)

type CsvReader struct {
	*csv.Reader
	columns        []*ColumnInfo
	skipJsonNull   bool
	headers        []string
	sliceDelimiter string
	lock           sync.Mutex
}

func NewCsvReader(r io.Reader, opts ...ReaderOption) *CsvReader {
	cr := &CsvReader{
		Reader:         csv.NewReader(r),
		sliceDelimiter: defaultSliceDelimiter,
	}
	for _, opt := range opts {
		opt(cr)
	}
	return cr
}

type ReaderOption func(*CsvReader)

func WithReaderSliceDelimiter(d string) ReaderOption {
	return func(cr *CsvReader) {
		cr.sliceDelimiter = d
	}
}

func WithReaderColumnInfos(infos []*ColumnInfo) ReaderOption {
	return func(cr *CsvReader) {
		cr.columns = infos
	}
}

func WithReaderSkipJSONNull(skip bool) ReaderOption {
	return func(cr *CsvReader) {
		cr.skipJsonNull = skip
	}
}

func (r *CsvReader) readValue() (map[string][]string, error) {
	if len(r.headers) == 0 {
		row, err := r.Read()
		if err != nil {
			return nil, err
		}

		// Filter unexpected characters.
		for i := 0; i < len(row); i++ {
			col := strings.TrimSpace(row[i])
			col = strings.TrimFunc(col, func(r rune) bool {
				return !unicode.IsGraphic(r)
			})
			row[i] = col
		}
		r.headers = row
	}

	row, err := r.Read()
	if err != nil {
		return nil, err
	}

	if len(row) != len(r.headers) {
		return nil, errors.New("Length of values not equals to length of header")
	}
	ret := make(map[string][]string)
	for i := 0; i < len(r.headers); i++ {
		ret[r.headers[i]] = append(ret[r.headers[i]], row[i])
	}
	return ret, nil
}

func (r *CsvReader) ReadAllStructs(i interface{}) error {
	val := reflect.ValueOf(i)
	if val.Kind() != reflect.Ptr {
		return errors.New("Input slice need to be a ptr")
	}
	if val.Elem().Kind() != reflect.Slice {
		return errors.New("Input need to be a slice")
	}
	typ := val.Elem().Type().Elem()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	for {
		inner := reflect.New(typ)
		if err := r.ReadStruct(inner.Interface()); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		val.Elem().Set(reflect.Append(val.Elem(), inner))
	}
}

func (r *CsvReader) ReadStruct(i interface{}) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.New("Input need to be a struct")
	}

	if len(r.columns) == 0 {
		r.columns = GenerateColumnInfos(val.Type())
	}

	values, err := r.readValue()
	if err != nil {
		return err
	}

	var allError error
	for _, col := range r.columns {
		if r.skipJsonNull && col.IsJsonNull {
			continue
		}
		value := values[col.HeaderName]
		if len(value) == 0 {
			// Is it possible?
			continue
		}
		v := val.FieldByName(col.LookupField)
		switch v.Kind() {
		case reflect.Invalid:
			// Nothing happen?
		case reflect.String:
			// Try to parse the string to correspond type.
			// NOTE: Using fmt.Sscanf("%v") will only parse the first
			// space-delimited token. If the cell contains like "123 abc",
			// only "123" will be parsed, while "abc" ignored.
			v.SetString(value[0])
		case reflect.Slice:
			if !col.IsSlice {
				allError = multierror.Append(allError, errors.Errorf("Field %s is not of kind slice?", col.LookupField))
				continue
			}
			var segs []string
			if col.NumSpan <= 1 {
				// Split by delimiter
				if value[0] != "" {
					segs = strings.Split(value[0], r.sliceDelimiter)
				}
			} else {
				segs = value
			}

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
		default:
			_, err := fmt.Sscanf(value[0], "%v", v.Addr().Interface())
			if err != nil && err != io.EOF {
				allError = multierror.Append(allError, err)
			}
		}
	}
	return allError
}

type FileCsvReader struct {
	*CsvReader
	File *os.File
}

func NewFileCsvReader(filename string) *FileCsvReader {
	file, err := os.Open(filename)
	if err != nil {
		goutils.LogError(err)
		return nil
	}
	return &FileCsvReader{
		CsvReader: NewCsvReader(file),
		File:      file,
	}
}

func (r *FileCsvReader) Close() error {
	return r.File.Close()
}
