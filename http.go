package goutils

import (
	"encoding/json"
	"io"
	"net/http"
)

func OutputHttpJson(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func OutputHttpOk(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "ok\n")
}
