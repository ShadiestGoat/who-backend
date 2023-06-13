package api

type HTTPError struct {
	Msg    string
	Status int
}

func (e HTTPError) Error() string {
	return e.Msg
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
	Msg: "You got bad http body",
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