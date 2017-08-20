package authentication

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/sessions"
	"github.com/matematik7/gongo"
	"github.com/matematik7/gongo/authorization"
	"github.com/matematik7/gongo/render"
)

type Authentication struct {
	authorization *authorization.Authorization
	render        *render.Render

	appURL string
}

func New(appURL string) *Authentication {
	auth := &Authentication{
		appURL: appURL,
	}

	return auth
}

func (auth *Authentication) ServeMux() http.Handler {
	router := chi.NewRouter()

	auth.ConfigureGothRoutes(router)

	return router
}

func (auth *Authentication) Configure(app gongo.App) error {
	auth.authorization = app["Authorization"].(*authorization.Authorization)
	auth.render = app["Render"].(*render.Render)
	auth.ConfigureGoth(app["Store"].(sessions.Store), auth.appURL)

	return nil
}
