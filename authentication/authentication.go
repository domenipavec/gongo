package authentication

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/sessions"
)

type Authentication struct {
	authorization Authorization
}

type Authorization interface {
	Login(w http.ResponseWriter, r *http.Request, id, name, email, avatarURL string) error
	Logout(w http.ResponseWriter, r *http.Request) error
}

func New() *Authentication {
	auth := &Authentication{}

	return auth
}

func (auth *Authentication) ServeMux() *chi.Mux {
	router := chi.NewRouter()

	auth.ConfigureGothRoutes(router)

	return router
}

func (auth *Authentication) Configure(store sessions.Store, authorization Authorization, appURL string) error {
	auth.authorization = authorization
	auth.ConfigureGoth(store, appURL)

	return nil
}
