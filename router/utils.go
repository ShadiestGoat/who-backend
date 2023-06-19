package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shadiestgoat/who/api"
)

type router struct {
	*chi.Mux
}

// unmarshals json from request body into p
// if the result is not ok, it is handled, and it returns true
// intended use:
// 
// if err := unmarshalNotOk(w, r, &body); err != nil {
// 		return nil, error
// }
func unmarshalNotOk(w http.ResponseWriter, r *http.Request, p any) error {
	if r.Body == nil {
		return api.ErrBadBody
	}

	err := json.NewDecoder(r.Body).Decode(p)
	if err != nil {
		return api.ErrBadBody
	}
	
	return nil
}

type handler func(w http.ResponseWriter, r *http.Request) (any, error)

func wResp(v any, w http.ResponseWriter) {
	json.NewEncoder(w).Encode(v)
}

func wRespErr(err error, w http.ResponseWriter) {
	var resp any

	if httpErr, ok := err.(api.HTTPErrorI); ok {
		w.WriteHeader(httpErr.StatusCode())
		resp = httpErr
	} else {
		w.WriteHeader(501)
		resp = &api.HTTPError{
			Msg: err.Error(),
		}
	}

	wResp(resp, w)
}

func wrap(h handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := h(w, r)
		if err != nil {
			wRespErr(err, w)
			return
		}

		wResp(resp, w)
	}
}

func (r *router) Get(path string, h handler) {
	r.Mux.Get(path, wrap(h))
}

func (r *router) Post(path string, h handler) {
	r.Mux.Post(path, wrap(h))
}

func (r *router) Put(path string, h handler) {
	r.Mux.Put(path, wrap(h))
}

func (r *router) Delete(path string, h handler) {
	r.Mux.Delete(path, wrap(h))
}

func (r *router) Patch(path string, h handler) {
	r.Mux.Patch(path, wrap(h))
}

func newRouter() *router {
	return &router{
		Mux: chi.NewRouter(),
	}
}
