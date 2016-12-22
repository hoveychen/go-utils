// Package presto provides struct declaration style wrapper to access presto result.
// See http://prestodb.io
// ----------------------------
// Example usage:
// type Info struct {
//     Name    string `presto:"name"`
//     Salary  int    `presto:"salary"`
//     Married bool   `presto:"is_married"`
// }
//
// p, err := presto.NewPrestoQuery(`Select * from employee`)
// if err != nil {
//    ...
// }
// defer p.Close()
// ret := Info{}
// for p.Next(&ret) {
//    ....
// }
package presto

import (
	"errors"
	"reflect"
	"strings"
	"time"

	go_presto "github.com/colinmarc/go-presto"
	"github.com/hoveychen/go-utils"
	"github.com/hoveychen/go-utils/flags"
)

var (
	defaultPrestoHost    = flags.String("prestoHost", "http://127.0.0.1:9997", "Presto server host address. Support multiple nodes with comma-delimited address.")
	defaultPrestoUser    = flags.String("prestoUser", "", "Default presto user.")
	defaultPrestoSource  = flags.String("prestoSource", "", "Default presto source")
	defaultPrestoCatalog = flags.String("prestoCatalog", "hive", "Default presto catalog to query")
	defaultPrestoSchema  = flags.String("prestoSchema", "event", "Default presto schema to query")

	defaultConfig *PrestoConfig
)

type PrestoQuery struct {
	query       *go_presto.Query
	columnIndex map[string]int
	sql         string
	alive       bool
}

type PrestoConfig struct {
	Host    string
	User    string
	Source  string
	Catalog string
	Schema  string

	pos int
}

func (cfg *PrestoConfig) GetPrestHost() string {
	candidates := strings.Split(strings.TrimSpace(cfg.Host), ",")
	cfg.pos = (cfg.pos + 1) % len(candidates)
	return candidates[cfg.pos]
}

func NewPrestoQueryWithConfig(sql string, cfg *PrestoConfig) (*PrestoQuery, error) {
	// TODO(yuheng): Introduce better way to manage the node pool.
	// Should always choose the alive node instead of death presto node.
	host := cfg.GetPrestHost()
	if host == "" {
		return nil, errors.New("No presto server host was set(missing --prestoHost)")
	}
	query, err := go_presto.NewQuery(host, cfg.User, cfg.Source, cfg.Catalog, cfg.Schema, sql)
	if err != nil {
		return nil, err
	}
	index := map[string]int{}
	for i, col := range query.Columns() {
		index[col] = i
	}
	q := &PrestoQuery{
		query:       query,
		columnIndex: index,
		sql:         sql,
		alive:       true,
	}
	return q, nil

}

func NewPrestoQuery(sql string) (*PrestoQuery, error) {
	if defaultConfig == nil {
		cfg := &PrestoConfig{}
		cfg.Host = *defaultPrestoHost
		cfg.User = *defaultPrestoUser
		cfg.Source = *defaultPrestoSource
		cfg.Catalog = *defaultPrestoCatalog
		cfg.Schema = *defaultPrestoSchema
		defaultConfig = cfg
	}
	return NewPrestoQueryWithConfig(sql, defaultConfig)
}

func (pq *PrestoQuery) Close() {
	if pq.alive {
		pq.alive = false
		pq.query.Close()
	}
}

func (pq *PrestoQuery) Next(result interface{}) bool {
	if pq.query == nil {
		goutils.LogError("Must invoke NewPrestoQuery() to initialize")
		return false
	}
	if !pq.alive {
		goutils.LogError("This query has been closed.")
		return false
	}
	val := reflect.ValueOf(result).Elem()
	data, err := pq.query.Next()
	if err != nil {
		goutils.LogError("Unexpected error", err)
		pq.Close()
		return false
	}
	if data == nil {
		pq.Close()
		return false
	}
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		tag := typeField.Tag.Get("presto")
		if tag == "" || tag == "-" {
			continue
		}
		loc, hit := pq.columnIndex[tag]
		if hit {
			if !val.Field(i).CanSet() {
				goutils.LogError(val.Field(i).Type().Name(), " can't be set.")
				return false
			}

			kind := val.Field(i).Kind()
			d := reflect.ValueOf(data[loc])

			if kind == reflect.Struct && typeField.Type.Name() == "Time" && typeField.Type.PkgPath() == "time" {
				v := int64(d.Float())
				val.Field(i).Set(reflect.ValueOf(time.Unix(v/1000, v%1000*1000000)))
			} else if kind == reflect.Int {
				val.Field(i).SetInt(int64(d.Float()))
			} else if kind == reflect.Float64 {
				val.Field(i).SetFloat(d.Float())
			} else if kind == reflect.String {
				val.Field(i).SetString(d.String())
			} else {
				goutils.LogError("Type not matched", typeField.Name, kind.String(), d.Kind().String())
				return false
			}
		} else {
			if val.Field(i).CanSet() {
				val.Field(i).Set(reflect.Zero(val.Field(i).Type()))
			}
		}
	}
	return true
}
