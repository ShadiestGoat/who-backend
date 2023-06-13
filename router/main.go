package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shadiestgoat/who/api"
)

type Router struct {
	*chi.Mux
}

type Handler func(w http.ResponseWriter, r *http.Request) (any, error)

func wResp(v any, w http.ResponseWriter) {
	json.NewEncoder(w).Encode(v)
}

func wrap(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := h(w, r)
		if err != nil {
			if httpErr, ok := err.(*api.HTTPError); ok {
				w.WriteHeader(httpErr.Status)
				resp = httpErr
			} else {
				w.WriteHeader(501)
				resp = &api.HTTPError{
					Msg: err.Error(),
				}
			}
		}

		wResp(resp, w)
	}
}

func (r *Router) Get(path string, h Handler) {
	r.Mux.Get(path, wrap(h))
}

func (r *Router) Post(path string, h Handler) {
	r.Mux.Post(path, wrap(h))
}

func (r *Router) Put(path string, h Handler) {
	r.Mux.Put(path, wrap(h))
}

func (r *Router) Delete(path string, h Handler) {
	r.Mux.Delete(path, wrap(h))
}

func (r *Router) Patch(path string, h Handler) {
	r.Mux.Patch(path, wrap(h))
}

func NewRouter() *Router {
	return &Router{
		Mux: chi.NewRouter(),
	}
}

func MainRouter() {
	r := NewRouter()

	r.Mount("/users", routerUsers())
}

func routerUsers() http.Handler {
	r := NewRouter()

	r.Get(`/users`, func(w http.ResponseWriter, r *http.Request) (any, error) {
		// decode stuff here, body etc
		username := chi.URLParam(r)

		return api.GetUser(username)
	})
}
