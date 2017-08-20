package render

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/flosch/pongo2"
	"github.com/pkg/errors"
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

type Render struct {
	isProd bool

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

	r.AddContextFunc(func(r *http.Request, ctx Context) {
		ctx["request"] = Request{
			Method: r.Method,
			Path:   r.URL.Path,
		}
	})

	return r
}

func (r *Render) AddTemplates(t Templates) {
	r.loader.Add(t)
}

func (r *Render) AddContextFunc(f ContextFunc) {
	r.contextFuncs = append(r.contextFuncs, f)
}

func (r *Render) Template(w http.ResponseWriter, req *http.Request, name string, ctx Context) {
	for _, cf := range r.contextFuncs {
		cf(req, ctx)
	}
	t, err := r.templateSet.FromCache(name)
	if err != nil {
		r.Error(w, req, err)
		return
	}
	err = t.ExecuteWriter(pongo2.Context(ctx), w)
	if err != nil {
		r.Error(w, req, err)
		return
	}
}

func (r *Render) Error(w http.ResponseWriter, req *http.Request, err error) {
	msg := err.Error()
	log.Println(err)
	if r.isProd {
		msg = "Sorry, something went wrong."
	}

	w.WriteHeader(http.StatusInternalServerError)
	r.Template(w, req, "error.html", Context{
		"title": "Server Error",
		"msg":   msg,
	})
}
