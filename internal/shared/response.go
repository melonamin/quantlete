package shared

import "encoding/json"

// Response is the standard API response wrapper for both HTTP and WASM.
// All endpoints return this structure for consistency.
type Response struct {
	OK      bool   `json:"ok"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse creates a successful response with data.
func SuccessResponse(data any) Response {
	return Response{OK: true, Data: data}
}

// SuccessMessage creates a successful response with a message (no data).
func SuccessMessage(msg string) Response {
	return Response{OK: true, Message: msg}
}

// ErrorResponse creates an error response.
func ErrorResponse(err error) Response {
	return Response{OK: false, Error: err.Error()}
}

// ErrorMessage creates an error response from a string message.
func ErrorMessage(msg string) Response {
	return Response{OK: false, Error: msg}
}

// ToJSON serializes the response to a JSON string.
// Used by WASM to return responses to JavaScript.
func (r Response) ToJSON() string {
	b, err := json.Marshal(r)
	if err != nil {
		return `{"ok":false,"error":"JSON marshal error"}`
	}
	return string(b)
}
