package authentication

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/matematik7/gongo"
)

type Authentication struct {
	authorization gongo.Authorization
	render        gongo.Render

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
	auth.authorization = app.Authorization
	auth.render = app.Render
	auth.ConfigureGoth(app.Store, auth.appURL)

	return nil
}

func (auth Authentication) Resources() []interface{} {
	return nil
}

func (auth Authentication) Name() string {
	return "Authentication"
}
