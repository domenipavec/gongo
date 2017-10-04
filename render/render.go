package render

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/sessions"
	"github.com/matematik7/gongo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Templates interface {
	Open(name string) (http.File, error)
}

type Context map[string]interface{}

type ContextFunc func(r *http.Request, ctx Context)

type templateLoader struct {
	templateSources []Templates
}

func (tl *templateLoader) Add(t Templates) {
	tl.templateSources = append(tl.templateSources, t)
}

func (tl templateLoader) Abs(base, name string) string {
	return name
}

func (tl templateLoader) Get(path string) (io.Reader, error) {
	for _, source := range tl.templateSources {
		f, err := source.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		return f, nil
	}

	return nil, errors.Errorf("template %s not found", path)
}

type FieldsProvider interface {
	LoggerFields(context.Context) map[string]interface{}
}

type Render struct {
	isProd bool

	log             *logrus.Logger
	fieldsProviders []FieldsProvider
	store           sessions.Store

	templateSet  *pongo2.TemplateSet
	loader       *templateLoader
	contextFuncs []ContextFunc
}

type Request struct {
	Method string
	Path   string
}

func New(isProd bool) *Render {
	loader := &templateLoader{}
	r := &Render{
		isProd:      isProd,
		templateSet: pongo2.NewSet("render-template-set", loader),
		loader:      loader,
	}

	r.templateSet.Debug = !isProd

	r.AddContextFunc(func(req *http.Request, ctx Context) {
		ctx["request"] = Request{
			Method: req.Method,
			Path:   req.URL.Path,
		}
	})

	return r
}

func (r *Render) Configure(app gongo.App) error {
	r.log = app["Log"].(*logrus.Logger)

	for _, component := range app {
		if fieldsProvider, ok := component.(FieldsProvider); ok {
			r.fieldsProviders = append(r.fieldsProviders, fieldsProvider)
		}
	}

	r.store = app["Store"].(sessions.Store)

	return nil
}

func (r *Render) AddTemplates(t Templates) {
	r.loader.Add(t)
}

func (r *Render) AddContextFunc(f ContextFunc) {
	r.contextFuncs = append(r.contextFuncs, f)
}

func (r *Render) AddFlash(w http.ResponseWriter, req *http.Request, flash interface{}) error {
	session, err := r.store.Get(req, "render")
	if err != nil {
		return errors.Wrap(err, "could not get store")
	}

	session.AddFlash(flash)

	if err := session.Save(req, w); err != nil {
		return errors.Wrap(err, "could not save session")
	}

	return nil
}

func (r *Render) renderTemplate(w http.ResponseWriter, req *http.Request, name string, ctx Context) error {
	if strings.HasSuffix(name, ".html") {
		w.Header().Set("Content-Type", "text/html")
	}
	for _, cf := range r.contextFuncs {
		cf(req, ctx)
	}

	session, err := r.store.Get(req, "render")
	if err != nil {
		return errors.Wrap(err, "could not get render store")
	}
	ctx["flashes"] = session.Flashes()
	err = session.Save(req, w)
	if err != nil {
		return errors.Wrap(err, "could not save render session")
	}

	t, err := r.templateSet.FromCache(name)
	if err != nil {
		return errors.Wrapf(err, "could not get template %s", name)
	}
	err = t.ExecuteWriter(pongo2.Context(ctx), w)
	if err != nil {
		// Do not log network errors (there were a lot of write on closed socket errors)
		if _, ok := err.(*net.OpError); !ok {
			return errors.Wrapf(err, "could not execute template %s", name)
		}
	}

	return nil
}

func (r *Render) Template(w http.ResponseWriter, req *http.Request, name string, ctx Context) {
	err := r.renderTemplate(w, req, name, ctx)
	if err != nil {
		r.Error(w, req, err)
	}
}

func (r *Render) NotFound(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	r.Template(w, req, "error.html", Context{
		"title": "Not Found",
		"msg":   "This is not the web page you are looking for.",
	})
}

func (r *Render) MethodNotAllowed(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	r.Template(w, req, "error.html", Context{
		"title": "Method Not Allowed",
		"msg":   "Your position's correct, except... not this method.",
	})
}

func (r *Render) Forbidden(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	r.Template(w, req, "error.html", Context{
		"title": "Forbidden",
		"msg":   "The Force is not strong with you.",
	})
}

func (r *Render) Error(w http.ResponseWriter, req *http.Request, err error) {
	msg := err.Error()

	fields := logrus.Fields{
		"RequestId": middleware.GetReqID(req.Context()),
		"URL":       req.URL.String(),
	}
	// TODO: separate this to logger package and use fields from main.go
	for _, fieldsProvider := range r.fieldsProviders {
		for k, v := range fieldsProvider.LoggerFields(req.Context()) {
			fields[k] = v
		}
	}

	r.log.WithFields(fields).Error(msg)

	if r.isProd {
		msg = "Sorry, something went wrong."
	}

	w.WriteHeader(http.StatusInternalServerError)
	// do not handle error, showing error template is best effort
	r.renderTemplate(w, req, "error.html", Context{
		"title": "Server Error",
		"msg":   msg,
	})
}
