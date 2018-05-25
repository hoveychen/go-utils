package goutils

import (
	"bytes"
	"text/template"

	"github.com/hoveychen/go-utils/gomap"
)

var (
	textTmplCache = gomap.New()
)

type Var map[string]interface{}

func Sprintt(textTmpl string, data interface{}) string {
	ret := textTmplCache.GetOrCreate(textTmpl, func() interface{} {
		tpl, err := template.New("test").Parse(textTmpl)
		if err != nil {
			LogError(err)
			return nil
		}
		return tpl
	})

	if ret == nil {
		// Not valid text template.
		return ""
	}

	tmpl := (ret).(*template.Template)
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		LogError(err)
		return ""
	}

	return buf.String()
}
