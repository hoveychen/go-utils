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

var bomUtf8 = []byte{0xEF, 0xBB, 0xBF}

// CsvWriter extends the encoding/csv writer, supporting writting struct, and
// shortcut to write to a file.
type CsvWriter struct {
	*csv.Writer
	w              io.Writer
	columns        []*ColumnInfo
	sliceDelimiter string
	skipJsonNull   bool
	headerWritten  bool
	appendBom      bool
	lock           sync.Mutex
}

func NewCsvWriter(w io.Writer, opts ...WriterOption) *CsvWriter {
	cw := &CsvWriter{
		Writer:         csv.NewWriter(w),
		w:              w,
		sliceDelimiter: defaultSliceDelimiter,
		skipJsonNull:   true,
		appendBom:      true,
	}
	for _, opt := range opts {
		opt(cw)
	}
	return cw
}

type WriterOption func(*CsvWriter)

func WithColumnInfos(infos []*ColumnInfo) WriterOption {
	return func(cw *CsvWriter) {
		cw.columns = infos
	}
}

func WithSliceDelimiter(d string) WriterOption {
	return func(cw *CsvWriter) {
		cw.sliceDelimiter = d
	}
}

func WithSkipJSONNull(skip bool) WriterOption {
	return func(cw *CsvWriter) {
		cw.skipJsonNull = skip
	}
}

func WithAppendBOM(enabled bool) WriterOption {
	return func(cw *CsvWriter) {
		cw.appendBom = enabled
	}
}

func (w *CsvWriter) writeHeader() error {
	if w.appendBom {
		if _, err := w.w.Write(bomUtf8); err != nil {
			return err
		}
	}
	var fields []string
	for _, col := range w.columns {
		if w.skipJsonNull && col.IsJsonNull {
			continue
		}
		for i := 0; i < col.NumSpan; i++ {
			fields = append(fields, col.HeaderName)
		}
	}
	return w.Write(fields)
}

func limitContent(str string, limit int) string {
	if limit <= 0 || len(str) <= limit {
		return str
	}
	return str[:limit]
}

func (w *CsvWriter) genSlice(col *ColumnInfo, v *reflect.Value) []string {
	var values []string
	if v.Kind() != reflect.Slice {
		values = append(values, fmt.Sprint(v.Interface()))
	} else {
		for i := 0; i < v.Len(); i++ {
			values = append(values, fmt.Sprint(v.Index(i).Interface()))
		}
	}

	if col.NumSpan == 1 {
		// Merge values into one cell with delimiter.
		str := strings.Join(values, w.sliceDelimiter)
		return []string{limitContent(str, col.Limit)}
	}

	if len(values) >= col.NumSpan {
		values = values[:col.NumSpan]
	}
	if len(values) < col.NumSpan {
		for i := len(values); i < col.NumSpan; i++ {
			values = append(values, "")
		}
	}

	for i := 0; i < len(values); i++ {
		values[i] = limitContent(values[i], col.Limit)
	}

	return values
}

func (w *CsvWriter) WriteStruct(i interface{}) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.New("Input need to be a struct")
	}

	if len(w.columns) == 0 {
		w.columns = GenerateColumnInfos(val.Type())
	}

	if !w.headerWritten {
		w.writeHeader()
		w.headerWritten = true
	}

	var out []string
	for _, col := range w.columns {
		if w.skipJsonNull && col.IsJsonNull {
			continue
		}
		v := val.FieldByName(col.LookupField)
		if col.IsSlice {
			out = append(out, w.genSlice(col, &v)...)
		} else {
			str := fmt.Sprint(v.Interface())
			out = append(out, limitContent(str, col.Limit))
		}
	}
	w.Write(out)
	return nil
}

func (w *CsvWriter) Close() error {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.Flush()
	return w.Error()
}

type FileCsvWriter struct {
	*CsvWriter
	File *os.File
}

// NewFileCsvWriter is combination of creating file and passing it to a writer.
func NewFileCsvWriter(filename string) *FileCsvWriter {
	file, err := os.Create(filename)
	if err != nil {
		goutils.LogError(err)
		return nil
	}
	file.Write(bomUtf8)
	return &FileCsvWriter{
		CsvWriter: NewCsvWriter(file),
		File:      file,
	}
}

func (fcw *FileCsvWriter) Close() error {
	fcw.Flush()
	return fcw.File.Close()
}
