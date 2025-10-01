package apperror

import "fmt"

type Error struct {
	Status  int
	Code    string
	Message string
	Details any
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(status int, code string, message string, details any) *Error {
	return &Error{Status: status, Code: code, Message: message, Details: details}
}

func NotFound(code string, message string) *Error {
	if code == "" {
		code = "not_found"
	}
	return New(404, code, message, nil)
}

func Conflict(code string, message string, details any) *Error {
	if code == "" {
		code = "conflict"
	}
	return New(409, code, message, details)
}

func BadRequest(code string, message string, details any) *Error {
	if code == "" {
		code = "bad_request"
	}
	return New(400, code, message, details)
}

func Forbidden(code string, message string) *Error {
	if code == "" {
		code = "forbidden"
	}
	return New(403, code, message, nil)
}

func Internal(message string, details any) *Error {
	return New(500, "internal_error", message, details)
}
