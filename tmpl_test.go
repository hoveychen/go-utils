package goutils

import (
	"testing"
)

func TestSprintt(t *testing.T) {
	type args struct {
		htmlTmpl string
		data     interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Normal",
			args{
				"Hello. My name is {{.Name}}, you can call me {{.Script}}",
				Var{
					"Name":   "Harry",
					"Script": "<i>H</i>",
				},
			},
			"Hello. My name is Harry, you can call me <i>H</i>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sprintt(tt.args.htmlTmpl, tt.args.data); got != tt.want {
				t.Errorf("Sprintt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSprintHTML(t *testing.T) {
	type args struct {
		htmlTmpl string
		data     interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Normal",
			args{"<p>{{.Hello}}</p>", Var{"Hello": "<a>"}},
			"<p>&lt;a&gt;</p>",
		},
		{
			"AsAttribute",
			args{"<a href='{{.Hello}}'></a>", Var{"Hello": "<a>"}},
			"<a href='%3ca%3e'></a>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SprintHTML(tt.args.htmlTmpl, tt.args.data); got != tt.want {
				t.Errorf("SprintHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}
