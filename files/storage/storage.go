package storage

import "io"

type Storage interface {
	URL(name string) (string, error)
	Save(name string, f io.Reader) error
	Delete(name string) error
	List(prefix string) ([]string, error)
}
