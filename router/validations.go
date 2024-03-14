package router

import (
	"github.com/grahms/godantic"
	"net/http"
)

type Validation struct {
	Exception Error
	godanic   godantic.Validate
}

func (va *Validation) contentType() (int, Error) {
	exp := Error{
		Code:    "CONTENT_TYPE_ERR",
		Reason:  "Unprocessable Entity",
		Message: "The request content type is not valid, content type should be `application/json`",
	}
	return http.StatusUnprocessableEntity, exp

}

func (va *Validation) ProcessorErr(perr *Err) (int, Error) {
	exp := Error{
		Code:    perr.ErrCode,
		Reason:  perr.ErrReason,
		Message: perr.Message,
	}
	return perr.StatusCode, exp

}

func (va *Validation) InputErr(err error) (int, Error) {
	if err, ok := err.(*godantic.Error); ok {
		return 400, Error{
			Reason:  BADREQUEST,
			Code:    err.ErrType,
			Message: err.Message}
	}
	return 0, Error{Message: err.Error()}
}
