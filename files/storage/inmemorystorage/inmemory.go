package inmemorystorage

import (
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

type InMemoryStorage struct {
	urlPrefix string

	mutex sync.RWMutex
	data  map[string][]byte
}

func New(urlPrefix string) *InMemoryStorage {
	return &InMemoryStorage{
		urlPrefix: urlPrefix,
		data:      make(map[string][]byte),
	}
}

func (ims *InMemoryStorage) URL(name string) (string, error) {
	return ims.urlPrefix + "/" + name, nil
}

func (ims *InMemoryStorage) Save(name string, f io.Reader) error {
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return errors.Wrap(err, "could not save")
	}

	ims.mutex.Lock()
	defer ims.mutex.Unlock()

	ims.data[name] = data

	return nil
}

func (ims *InMemoryStorage) Delete(name string) error {
	ims.mutex.Lock()
	defer ims.mutex.Unlock()

	delete(ims.data, name)

	return nil
}

func (ims *InMemoryStorage) ServeMux() http.Handler {
	router := chi.NewRouter()

	router.Get("/{filename}", func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "filename")

		ims.mutex.RLock()
		defer ims.mutex.RUnlock()

		data, ok := ims.data[name]
		if !ok {
			// TODO: render not found
			http.NotFound(w, r)
			return
		}

		w.Write(data)
	})

	return router
}
