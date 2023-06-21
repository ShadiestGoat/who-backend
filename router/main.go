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
	r.Mount(`/questions/{id}`, routerQuestions())
	r.Mount(`/auth`, routerAuth())

	return r
}

type reqNewQuiz struct {
	Quiz      api.Quiz        `json:"quiz"`
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

		body.Quiz.AuthorID = r.Context().Value(CTX_USER).(string)

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

		body.ID = chi.URLParam(r, "id")

		return api.EditQuiz(&body)
	})

	r.Delete("/", func(w http.ResponseWriter, r *http.Request) (any, error) {
		return api.DeleteQuiz(chi.URLParam(r, "id"))
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) (any, error) {
		return api.GetQuiz(chi.URLParam(r, "id"))
	})

	r.Get(`/questions`, func(w http.ResponseWriter, r *http.Request) (any, error) {
		return api.GetQuestions(chi.URLParam(r, "id"))
	})

	return r
}

type respPreview struct {
	Question1 *api.Question `json:"question"`
	Title     string `json:"title"`
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

		q, err := api.GetQuizFirstQuestion(quizID)

		if err != nil {
			return nil, err
		}

		resp := &respPreview{
			Question1: q,
			Title:     "Who the fuck is " + strings.ToUpper(chosenName[:1]) + chosenName[1:],
		}

		return resp, nil
	})

	return r
}

type reqAnswer struct {
	Answer string `json:"answer"`
}

// /questions/{id}
func routerQuestions() http.Handler {
	r := newRouter()

	r.Use(middlewareQuestion)

	r.Post(`/answer`, func(w http.ResponseWriter, r *http.Request) (any, error) {
		body := reqAnswer{}

		if err := unmarshalNotOk(w, r, &body); err != nil {
			return nil, err
		}

		return api.AnswerQuestion(chi.URLParam(r, `id`), body.Answer)
	})

	r.With(middlewareAuth).With(middlewareQuestionAuth).Post(`/`, wrap(func(w http.ResponseWriter, r *http.Request) (any, error) {
		body := api.FullQuestion{}

		if err := unmarshalNotOk(w, r, &body); err != nil {
			return nil, err
		}
		
		body.ID = chi.URLParam(r, "id")

		return api.EditQuestion(&body)
	}))

	return r
}

type reqAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type respAuth struct {
	ID string `json:"id"`
	Token string `json:"token"`
}

func routerAuth() http.Handler {
	r := newRouter()

	r.Post(`/`, func(w http.ResponseWriter, r *http.Request) (any, error) {
		body := reqAuth{}

		if err := unmarshalNotOk(w, r, &body); err != nil {
			return nil, err
		}

		id, token, err := api.Exchange(body.Username, body.Password)

		return &respAuth{
			ID:    id,
			Token: token,
		}, err
	})

	return r
}
