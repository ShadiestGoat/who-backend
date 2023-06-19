package api

import (
	"strings"
)

type HTTPErrorI interface {
	StatusCode() int
}

type HTTPError struct {
	Msg    string `json:"error"`
	Status int `json:"-"`
}

func (e HTTPError) Error() string {
	return e.Msg
}

func (e HTTPError) StatusCode() int {
	return e.Status
}

func newHTTPErrorStack(errors []error) *HTTPErrorStack {
	code := -1
	msgs := []string{}

	for _, err := range errors {
		if err == nil {
			continue
		}

		if err, ok := err.(*HTTPError); ok && code == -1 {
			code = err.Status
		}

		msgs = append(msgs, err.Error())
	}

	if len(msgs) == 0 {
		return nil
	}
	
	if code == -1 {
		code = 500
	}

	return &HTTPErrorStack{
		Errors: msgs,
		Status: code,
	}
}

type HTTPErrorStack struct {
	Errors []string `json:"error"`
	Status int
}

func (e HTTPErrorStack) Error() string {
	return strings.Join(e.Errors, "\n")
}

func (e HTTPErrorStack) StatusCode() int {
	return e.Status
}

var ErrServerErr = &HTTPError{
	Msg:    "Server error! Could not handle it :(",
	Status: 500,
}

var ErrNotFound = &HTTPError{
	Msg:    "Resource doesn't exist",
	Status: 404,
}

var ErrBadBody = &HTTPError{
	Msg:    "You got bad http body",
	Status: 400,
}

var ErrUniqueUname = &HTTPError{
	Msg:    "This username is not unique!",
	Status: 401,
}

var ErrNoAuth = &HTTPError{
	Msg:    "You are not authorized",
	Status: 401,
}

var ErrBadName = &HTTPError{
	Msg: "Your name is not acceptable",
	Status: 400,
}
