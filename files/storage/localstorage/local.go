package localstorage

import (
	"io"
	"net/http"
	"os"
	"path"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

type LocalStorage struct {
	urlPrefix string
	folder    string
}

func New(folder, urlPrefix string) (*LocalStorage, error) {
	err := os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "could not create storage folder")
	}

	return &LocalStorage{
		urlPrefix: urlPrefix,
		folder:    folder,
	}, nil
}

func (ls *LocalStorage) URL(name string) (string, error) {
	return ls.urlPrefix + "/" + name, nil
}

func (ls LocalStorage) getPathName(name string) string {
	return path.Join(ls.folder, name)
}

func (ls *LocalStorage) Save(name string, input io.Reader) error {
	pathName := ls.getPathName(name)

	f, err := os.Create(pathName)
	if err != nil {
		return errors.Wrapf(err, "could not create file %s", pathName)
	}
	defer f.Close()

	_, err = io.Copy(f, input)
	if err != nil {
		return errors.Wrap(err, "could not copy to file")
	}

	return nil
}

func (ls *LocalStorage) Delete(name string) error {
	pathName := ls.getPathName(name)

	if err := os.Remove(pathName); err != nil {
		return errors.Wrapf(err, "could not delete file %s", pathName)
	}

	return nil
}

func (ls *LocalStorage) ServeMux() http.Handler {
	router := chi.NewRouter()

	router.Get("/{filename}", func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "filename")
		pathName := ls.getPathName(name)

		http.ServeFile(w, r, pathName)
	})

	return router
}
