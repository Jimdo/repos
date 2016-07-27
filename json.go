package main

import (
	"encoding/json"
	"net/http"
)

type apiError struct {
	Error string `json:"error"`
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		jsonError(w, err)
	}
}

func jsonError(w http.ResponseWriter, err error) {
	apiErr := apiError{err.Error()}
	jsonResponse(w, apiErr)
}
