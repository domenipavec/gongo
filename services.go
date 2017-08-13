package gongo

import (
	"net/http"
)

type Controller interface {
	Configure(app App) error
	Name() string
	Resources() []interface{}
	ServeMux() http.Handler
}

type Authentication interface {
	Controller
}

type Authorization interface {
	Controller

	Login(w http.ResponseWriter, r *http.Request, id, name, email, avatarURL string) error
	Logout(w http.ResponseWriter, r *http.Request) error
	AddPermission(code, name string) error
}

type Resources interface {
	Controller

	Register(group string, models ...interface{}) error
}

type Templates interface {
	Open(name string) (http.File, error)
}

type Context map[string]interface{}

type ContextFunc func(r *http.Request, ctx Context)

type Render interface {
	AddTemplates(Templates)
	AddContextFunc(ContextFunc)
	Template(w http.ResponseWriter, r *http.Request, name string, ctx Context)
	Error(w http.ResponseWriter, r *http.Request, err error)
}
