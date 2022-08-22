package csv

import (
	"reflect"
	"strconv"
	"strings"
)

type ColumnInfo struct {
	// Name in the csv header.
	HeaderName string
	// Field to lookup values in struct.
	LookupField string
	// Number of cells for this column. Applied when IsSlice = true
	NumSpan int
	// Max length of content in this column
	Limit      int
	IsSlice    bool
	IsJsonNull bool
}

func extractKV(s string) (k, v string) {
	if !strings.Contains(s, "=") {
		return s, ""
	}
	segs := strings.SplitN(s, "=", 2)
	return segs[0], segs[1]
}

func GenerateColumnInfos(typ reflect.Type) []*ColumnInfo {
	var ret []*ColumnInfo
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			// Unexported field will have PkgPath.
			continue
		}
		tag := field.Tag.Get("csv")

		var name string
		var ignore bool
		var numSpan int
		var limit int
		var isJsonNull bool
		var isSlice bool

		if field.Tag.Get("json") == "-" {
			isJsonNull = true
		}

		if field.Type.Kind() == reflect.Slice {
			isSlice = true
		}

		for _, seg := range strings.Split(tag, ",") {
			if seg == "-" {
				ignore = true
				continue
			}
			k, v := extractKV(seg)
			if v == "" {
				name = k
				continue
			}
			if k == "span" {
				num, _ := strconv.Atoi(v)
				if num >= 1 {
					numSpan = num
				}
			}
			if k == "limit" {
				num, _ := strconv.Atoi(v)
				if num >= 1 {
					limit = num
				}
			}
		}

		if ignore {
			continue
		}
		if name == "" {
			name = field.Name
		}
		// Only slice have multiple spans
		if numSpan <= 0 || !isSlice {
			numSpan = 1
		}
		if limit < 0 {
			continue
		}

		ret = append(ret, &ColumnInfo{
			HeaderName:  name,
			LookupField: field.Name,
			NumSpan:     numSpan,
			Limit:       limit,
			IsSlice:     isSlice,
			IsJsonNull:  isJsonNull,
		})
	}
	return ret
}
