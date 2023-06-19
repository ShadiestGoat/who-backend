package router

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shadiestgoat/who/api"
	"github.com/shadiestgoat/who/db"
)

type ctx int

const (
	CTX_AUTHOR ctx = iota
)

func middlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := ""

		err := db.QueryRowID(`SELECT id FROM users WHERE token = $1`, r.Header.Get("Authorization"), &id)

		if err != nil {
			if db.NoRows(err) {
				err = api.ErrNoAuth
			} else {
				err = api.ErrServerErr
			}

			wRespErr(err, w)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CTX_AUTHOR, id)))
	})
}

func middlewareQuiz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		author := ""

		err := db.QueryRowID(`SELECT author FROM quiz WHERE id = $1`, chi.URLParam(r, "id"), &author)

		if err != nil {
			wRespErr(api.ErrDBHandle(err), w)
			return
		}

		if author != r.Context().Value(CTX_AUTHOR).(string) {
			wRespErr(api.ErrNoAuth, w)
			return
		}

		next.ServeHTTP(w, r)
	})
}
