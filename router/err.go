package router

import "net/http"

// Error represents an error that occurred while processing an HTTP request.
type Error struct {
	// Code is a unique identifier for the error.
	Code string `json:"code"`

	// Reason is a brief description of the error.
	Reason string `json:"reason"`

	// Message is a detailed explanation of the error.
	Message string `json:"message"`
}

// BADREQUEST is a constant that represents a bad request error.
const (
	BADREQUEST = "Bad Request"
)

// InvalidBody returns an HTTP status code and an Error representing an error
// where the input body is invalid.
func (e *Error) InvalidBody() (int, Error) {
	exp := Error{
		Code:    "INVALID_BODY_ERROR",
		Message: "The input body is invalid",
		Reason:  BADREQUEST,
	}
	return http.StatusBadRequest, exp
}

// EmptyBody returns an HTTP status code and an Error representing an error
// where the input body is empty.
func (e *Error) EmptyBody() (int, Error) {
	exp := Error{
		Code:    "EMPTY_BODY_ERROR",
		Message: "Empty body",
		Reason:  BADREQUEST,
	}
	return http.StatusBadRequest, exp
}

// InvalidContentType returns an HTTP status code and an Error representing
// an error where the input body has an invalid content type.
func (e *Error) InvalidContentType() (int, Error) {
	exp := Error{
		Code:    "COMTENT_TYPE_ERROR",
		Message: "Content type is not application/json",
		Reason:  BADREQUEST,
	}
	return http.StatusUnprocessableEntity, exp
}

func (e *Error) InternalServerError() (int, Error) {
	exp := Error{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "Internal Server Error",
		Reason:  "Something wrong happened in the backend system",
	}
	return http.StatusInternalServerError, exp
}

func (e *Error) MethodNotAllowed() (int, Error) {
	exp := Error{
		Code:    "METHOD_NOT_ALLOWED_ERROR",
		Message: "Method not Allowed",
		Reason:  "Resource does not allow this method",
	}
	return http.StatusMethodNotAllowed, exp
}

func (e *Error) ResourceNotFound() (int, Error) {
	exp := Error{
		Code:    "NOT_FOUND_ERROR",
		Message: "Not found",
		Reason:  "Resource not found",
	}
	return http.StatusNotFound, exp
}
