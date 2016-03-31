package utils

import (
	"encoding/json"
	"net/http"
)

type ResponseWrapper struct {
	ErrorCode    int         `json:"errcode"`
	ErrorMessage string      `json:"errmsg"`
	Data         interface{} `json:"data"`
}

func WriteJson(w http.ResponseWriter, data interface{}) error {
	resp := ResponseWrapper{ErrorCode: 0, Data: data}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		return err
	}
	w.Header().Add("content-type", "application/json")
	return nil
}

func WriteError(w http.ResponseWriter, errorCode int, errorMessage string) error {
	resp := ResponseWrapper{ErrorCode: errorCode, ErrorMessage: errorMessage}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		return err
	}

	w.Header().Add("content-type", "application/json")
	return nil
}

const (
	DateLayout = "2006/01/02 15:04:05"
)
