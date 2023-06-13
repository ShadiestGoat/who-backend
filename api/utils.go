package api

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/shadiestgoat/who/db"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// this will both clean a string & check for the correct length values.
// Both length args are non-inclusive, so minLength=0 means that the string has to have at least 1 byte
// If either length argument is < 0, the argument is not used.
// Returns if the string is not ok, intended use is:
//
// if err := cleanString(&s, ...); err != nil {return err}
func cleanString(s *string, minLength, maxLength int, key string) error {
	*s = strings.TrimSpace(*s)
	l := len(*s)

	if minLength >= 0 && l < minLength {
		return &HTTPError{
			Msg:    fmt.Sprintf("Key '%s' is too short", key),
			Status: 400,
		}
	}

	if maxLength >= 0 && l > maxLength {
		return &HTTPError{
			Msg:    fmt.Sprintf("Key '%s' is too long", key),
			Status: 400,
		}
	}

	return nil
}

func cleanStringArr(inp []string, minLength, maxLength int, key string) ([]string, error) {
	for i := range inp {
		err := cleanString(&inp[i], minLength, maxLength, key)
		if err != nil {
			return nil, err
		}
	}

	return inp, nil
}

func Capitalize(s string) string {
	return cases.Title(language.AmericanEnglish).String(s)
}

func ErrDBHandle(err error) *HTTPError {
	if db.NoRows(err) {
		return ErrNotFound
	}
	return ErrServerErr
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randGoodString(l int) string {
	b := make([]byte, l)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return string(b)
}
