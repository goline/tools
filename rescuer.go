package utils

import (
	"github.com/goline/lapi"
	"net/http"
)

func NewRescuer(logger lapi.Logger) lapi.Rescuer {
	return &FactoryRescuer{logger}
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type FactoryRescuer struct {
	logger lapi.Logger
}

func (r *FactoryRescuer) Rescue(connection lapi.Connection, err error) error {
	if connection == nil {
		return err
	}
	switch e := err.(type) {
	case lapi.SystemError:
		r.handleSystemError(connection, e)
	case lapi.StackError:
		r.handleStackError(connection, e)
	default:
		r.handleUnknownError(connection, e)
	}
	r.logger.Error("%v", err)

	return nil
}

func (r *FactoryRescuer) handleSystemError(c lapi.Connection, err lapi.SystemError) {
	switch err.Code() {
	case lapi.ERROR_HTTP_NOT_FOUND:
		c.Response().WithStatus(http.StatusNotFound).
			WithContent(&ErrorResponse{"ERROR_HTTP_NOT_FOUND", http.StatusText(http.StatusNotFound)})
	case lapi.ERROR_HTTP_BAD_REQUEST:
		c.Response().WithStatus(http.StatusBadRequest).
			WithContent(&ErrorResponse{"ERROR_HTTP_BAD_REQUEST", http.StatusText(http.StatusBadRequest)})
	default:
		c.Response().WithStatus(http.StatusInternalServerError).
			WithContent(&ErrorResponse{"ERROR_INTERNAL_SERVER_ERROR", err.Error()})
	}
}

func (r *FactoryRescuer) handleStackError(c lapi.Connection, err lapi.StackError) {
	c.Response().WithStatus(err.Status()).WithContent(&ErrorResponse{"", err.Error()})
}

func (r *FactoryRescuer) handleUnknownError(c lapi.Connection, err error) {
	if e, ok := err.(lapi.ErrorStatus); ok == true {
		c.Response().WithStatus(e.Status())
	} else {
		c.Response().WithStatus(http.StatusInternalServerError)
	}
	code := "ERROR_UNKNOWN_ERROR"
	if e, ok := err.(lapi.ErrorCoder); ok == true {
		code = e.Code()
	}
	c.Response().WithContent(&ErrorResponse{code, err.Error()})
}