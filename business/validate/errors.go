package validate

import (
	"encoding/json"
	"errors"
)

type ErrorResponse struct {
	Fields string `json:"fields,omitempty"`
	Error  string `json:"error"`
}

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type FieldErrors []FieldError

func (fe FieldErrors) Error() string {
	d, err := json.Marshal(fe)
	if err != nil {
		return err.Error()
	}
	return string(d)
}

func IsFieldErrors(err error) bool {
	var fe FieldErrors
	return errors.As(err, &fe)
}

type RequestError struct {
	Err    error
	Status int
	Fields error
}

func NewRequestError(err error, status int) error {
	return &RequestError{err, status, nil}
}

func (re RequestError) Error() string {
	return re.Err.Error()
}

func IsRequestError(err error) bool {
	var re *RequestError
	return errors.As(err, &re)
}

func GetRequestError(err error) *RequestError {
	var re *RequestError
	if !errors.As(err, &re) {
		return nil
	}
	return re
}
