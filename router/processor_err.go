package router

import "errors"

type ProcessorError struct {
	StatusCode int
	ErrCode    string
	ErrReason  string
	Message    string
	err        error
}

func (e *ProcessorError) Error() string {
	e.err = errors.New(e.Message)
	return e.err.Error()
}
