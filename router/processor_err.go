package router

import "errors"

type Err struct {
	StatusCode int
	ErrCode    string
	ErrReason  string
	Message    string
	err        error
}

func (e *Err) Error() string {
	e.err = errors.New(e.Message)
	return e.err.Error()
}
