package http

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error"`
}

type Error struct {
	Code    string      `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(response)
}

func Success(w http.ResponseWriter, status int, message string, data interface{}) {
	JSON(w, status, Response{
		Success: true,
		Message: message,
		Data:    data,
		Error:   nil,
	})
}

func ErrorResponse(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
	details interface{},
) {
	JSON(w, status, Response{
		Success: false,
		Message: message,
		Data:    nil,
		Error: Error{
			Code:    code,
			Details: details,
		},
	})
}
