package main

import (
	"net/http"

	"github.com/goccy/go-json"
)

type (
	ApiResponse struct {
		StatCode    int         `json:"stat_code"`
		StatMessage string      `json:"stat_message,omitempty"`
		ErrMessage  string      `json:"err_message,omitempty"`
		Data        interface{} `json:"data,omitempty"`
	}
)

func APIResponse(w http.ResponseWriter, req *ApiResponse) {
	reqByte, err := json.Marshal(req)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(req.StatCode)
	w.Write(reqByte)
}
