package router

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessorError_Error(t *testing.T) {
	err := &ProcessorError{
		StatusCode: 500,
		ErrCode:    "INTERNAL_ERROR",
		ErrReason:  "An internal error occurred while processing the request",
		Message:    "There was a problem processing your request. Please try again later.",
	}

	assert.Equal(t, "There was a problem processing your request. Please try again later.", err.Error())
	assert.Equal(t, "There was a problem processing your request. Please try again later.", err.err.Error())
}
