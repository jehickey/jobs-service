package main

import (
	"time"
)

type APIResponse struct {
	Success   bool        `json:"success"`
	Timestamp int64       `json:"timestamp"`
	Version   int         `json:"version"`
	Errors    []APIError  `json:"errors,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Creates a new prepopulated APIResponse object
func NewAPIResponse() APIResponse {
	return APIResponse{
		Success:   true,
		Timestamp: time.Now().UnixMilli(),
		Version:   1,
	}
}

// Helper to add errors to APIResponse
//
// OK defines if the call is ok to continue.  False = fail
func (r *APIResponse) AddError(ok bool, msg string) {
	r.Errors = append(r.Errors, APIError{Message: msg})
	if !ok {
		r.Success = false
	}
}
