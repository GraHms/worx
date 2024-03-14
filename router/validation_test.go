package router

import (
	"errors"
	"github.com/grahms/godantic"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestShouldCheckEnums(t *testing.T) {
	type MyStruct struct {
		Email *string `json:"email" binding:"required" enum:"data@mail.com,another@me.com"`
		Age   *int64  `binding:"required"`
	}
	validate := godantic.Validate{}
	wrongEmail := "iam@wrong.com"
	err := validate.InspectStruct(MyStruct{
		Email: &wrongEmail,
		Age:   new(int64),
	})
	validation := Validation{}

	validationErrs := err.(*godantic.Error)
	statusCode, exception := validation.InputErr(validationErrs)
	assert.Equal(t, 400, statusCode)
	assert.Equal(t, "INVALID_ENUM_ERR", exception.Code)
	assert.Equal(t, "Bad Request", exception.Reason)
	assert.Equal(t, "The field <email> must have one of the following values: data@mail.com, another@me.com, 'iam@wrong.com' was given", exception.Message)

}
func TestInput(t *testing.T) {
	validation := Validation{}

	type MyStruct struct {
		Email *string `json:"email" binding:"required"`
		Age   *int64  `binding:"required"`
	}
	validate := godantic.Validate{}

	err := validate.InspectStruct(MyStruct{})

	validationErrs := err.(*godantic.Error)

	statusCode, exception := validation.InputErr(validationErrs)
	assert.Equal(t, 400, statusCode)
	assert.Equal(t, "REQUIRED_FIELD_ERR", exception.Code)
	assert.Equal(t, "Bad Request", exception.Reason)
	assert.Equal(t, "The field <email> is required", exception.Message)

}

func TestProcessorErr(t *testing.T) {
	validation := Validation{}

	// Test with a Err struct
	perr := Err{
		ErrCode:    "123",
		ErrReason:  "Test Error",
		Message:    "This is a test error message",
		StatusCode: 500,
	}
	statusCode, exception := validation.ProcessorErr(&perr)
	assert.Equal(t, 500, statusCode)
	assert.Equal(t, "123", exception.Code)
	assert.Equal(t, "Test Error", exception.Reason)
	assert.Equal(t, "This is a test error message", exception.Message)
}

func TestValidation_Input_GodanticError(t *testing.T) {
	va := Validation{}

	// HandleCreate a GodanticError instance
	err := godantic.Error{}

	// Call the InputErr method with the GodanticError instance
	statusCode, exp := va.InputErr(&err)

	// Assert that the returned HTTP status code is 400
	assert.Equal(t, http.StatusBadRequest, statusCode)
	// Assert that the returned Error has the expected error code and message
	assert.Equal(t, "", exp.Code)
	assert.Equal(t, "Bad Request", exp.Reason)
	assert.Equal(t, err.Error(), exp.Message)
}

func TestShouldValidateNoneGodanticError(t *testing.T) {
	va := Validation{}

	// HandleCreate a GodanticError instance
	err := errors.New("some strange error")

	// Call the InputErr method with the GodanticError instance
	statusCode, exp := va.InputErr(err)

	// Assert that the returned HTTP status code is 400
	assert.Equal(t, 0, statusCode)
	// Assert that the returned Error has the expected error code and message

	assert.Equal(t, err.Error(), exp.Message)
}
