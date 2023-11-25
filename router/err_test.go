package router

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestExceptionEmptyBody(t *testing.T) {
	exp := Error{}
	code, err := exp.EmptyBody()
	assert.Equal(t, 400, code)
	assert.Equal(t, "Bad Request", err.Reason)
}

func TestExceptionInvalidBody(t *testing.T) {
	exp := Error{}
	code, _ := exp.InvalidBody()
	assert.Equal(t, 400, code)

}

func TestInvalidBody(t *testing.T) {
	exp := Error{}
	status, err := exp.InvalidBody()

	if status != http.StatusBadRequest {
		t.Errorf("Expected HTTP status code %d, got %d", http.StatusBadRequest, status)
	}

	if err.Code != "INVALID_BODY_ERROR" {
		t.Errorf("Expected error code %q, got %q", "001", err.Code)
	}

	if err.Reason != BADREQUEST {
		t.Errorf("Expected error reason %q, got %q", BADREQUEST, err.Reason)
	}

	if err.Message != "The input body is invalid" {
		t.Errorf("Expected error message %q, got %q", "The input body is invalid", err.Message)
	}
}

func TestEmptyBody(t *testing.T) {
	exp := Error{}
	status, err := exp.EmptyBody()

	if status != http.StatusBadRequest {
		t.Errorf("Expected HTTP status code %d, got %d", http.StatusBadRequest, status)
	}
	assert.Equal(t, http.StatusBadRequest, status)

	if err.Code != "EMPTY_BODY_ERROR" {
		t.Errorf("Expected error code %q, got %q", "002", err.Code)
	}

	if err.Reason != BADREQUEST {
		t.Errorf("Expected error reason %q, got %q", BADREQUEST, err.Reason)
	}

	if err.Message != "Empty body" {
		t.Errorf("Expected error message %q, got %q", "Empty body", err.Message)
	}
}

func TestInterServerError(t *testing.T) {
	e := Error{}

	status, err := e.InternalServerError()
	if status != http.StatusInternalServerError {
		t.Errorf("Expected HTTP status code %d, got %d", http.StatusInternalServerError, status)
	}
	assert.Equal(t, http.StatusInternalServerError, status)
	if err.Code != "INTERNAL_SERVER_ERROR" {
		t.Errorf("Expected error code '003', got '%s'", err.Code)
	}

	if err.Message != "Internal Server Error" {
		t.Errorf("Expected error message 'Internal Server Error', got '%s'", err.Message)
	}

	if err.Reason != "Something wrong happened in the backend system" {
		t.Errorf("Expected error reason 'Something wrong happened in the backend system', got '%s'", err.Reason)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	e := Error{}

	status, err := e.MethodNotAllowed()
	if status != http.StatusMethodNotAllowed {
		t.Errorf("Expected HTTP status code %d, got %d", http.StatusMethodNotAllowed, status)
	}
	assert.Equal(t, http.StatusMethodNotAllowed, status)
	if err.Code != "METHOD_NOT_ALLOWED_ERROR" {
		t.Errorf("Expected error code '004', got '%s'", err.Code)
	}

	if err.Message != "Method not Allowed" {
		t.Errorf("Expected error message 'Method not Allowed', got '%s'", err.Message)
	}

}

func TestInvalidContentType(t *testing.T) {
	// HandleCreate a new Error.
	e := &Error{}

	// Call the InvalidContentType method and save the returned values.
	statusCode, exp := e.InvalidContentType()

	// Verify that the returned HTTP status code is correct.
	if statusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected HTTP status code %d, got %d", http.StatusUnprocessableEntity, statusCode)
	}
	assert.Equal(t, statusCode, http.StatusUnprocessableEntity)

	// Verify that the returned Error has the correct code.
	if exp.Code != "COMTENT_TYPE_ERROR" {
		t.Errorf("Expected code %s, got %s", "003", exp.Code)
	}

	// Verify that the returned Error has the correct message.
	if exp.Message != "Content type is not application/json" {
		t.Errorf("Expected message %s, got %s", "Body cannot be empty", exp.Message)
	}

	// Verify that the returned Error has the correct reason.
	if exp.Reason != BADREQUEST {
		t.Errorf("Expected reason %s, got %s", BADREQUEST, exp.Reason)
	}
}

func TestResourceNotFound(t *testing.T) {
	// HandleCreate a new Error.
	e := &Error{}

	// Call the ResourceNotFound method and save the returned values.
	statusCode, exp := e.ResourceNotFound()

	// Use to assert.Equal method to verify that the returned HTTP status code is correct.
	assert.Equal(t, http.StatusNotFound, statusCode, "Expected HTTP status code %d, got %d", http.StatusNotFound, statusCode)

	// Use to assert.Equal method to verify that the returned Error has the correct code.
	assert.Equal(t, "NOT_FOUND_ERROR", exp.Code, "Expected code %s, got %s", "005", exp.Code)

	// Use to assert.Equal method to verify that the returned Error has the correct message.
	assert.Equal(t, "Not found", exp.Message, "Expected message %s, got %s", "Not found", exp.Message)

	// Use to assert.Equal method to verify that the returned Error has the correct reason.
	assert.Equal(t, "Resource not found", exp.Reason, "Expected reason %s, got %s", "Resource not found", exp.Reason)
}
