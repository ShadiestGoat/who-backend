package router

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/shadiestgoat/who/api"
	"github.com/shadiestgoat/who/db"
)

func MainRouter() http.Handler {
	r := newRouter()

	r.Mount(`/quizzes`, routerQuizzes())
	r.Mount(`/previews`, routerPreview())

	return r
}


type reqNewQuiz struct {
	Quiz api.Quiz `json:"quiz"`
	Questions []*api.Question `json:"questions"`	
}

func routerQuizzes() http.Handler {
	r := newRouter()

	r.Use(middlewareAuth)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) (any, error) {
		body := &reqNewQuiz{}

		if err := unmarshalNotOk(w, r, &body); err != nil {
			return nil, err
		}

		body.Quiz.AuthorID = r.Context().Value(CTX_AUTHOR).(string)

		return api.NewQuiz(&body.Quiz, body.Questions)
	})

	r.Handle("/{id}", routerQuizID())

	return r
}

func routerQuizID() http.Handler {
	r := newRouter()

	r.Use(middlewareQuiz)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) (any, error) {
		body := api.Quiz{}

		if err := unmarshalNotOk(w, r, &body); err != nil {
			return nil, err
		}

		body.AuthorID = r.Context().Value(CTX_AUTHOR).(string)
		body.ID = chi.URLParam(r, "id")

		return api.EditQuiz(&body)
	})

	r.Delete("/", func(w http.ResponseWriter, r *http.Request) (any, error) {
		return api.DeleteQuiz(chi.URLParam(r, "id"))
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) (any, error) {
		return api.GetQuiz(chi.URLParam(r, "id"))
	})

	return r
}

type respPreview struct {
	Question1 *api.Question
	Title     string
}

func routerPreview() http.Handler {
	r := newRouter()

	r.Get(`/{id}`, func(w http.ResponseWriter, r *http.Request) (any, error) {
		quizID := chi.URLParam(r, "id")

		chosenName := ""

		err := db.QueryRowID(`SELECT chosenname[1] FROM quiz WHERE id = $1`, quizID, &chosenName)

		if err != nil {
			return nil, api.ErrDBHandle(err)
		}

		resp := &respPreview{
			Question1: &api.Question{},
			Title:     "Who the fuck is " + strings.ToUpper(chosenName[:1]) + chosenName[1:],
		}

		return 
	})
}

/*

GET  /preview/{id}
POST /questions/{id}  

POST /auth *

*/
// Get
// New
// Edit Quiz (POST)
// Delete
// First question
// Answer