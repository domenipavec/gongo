package files

import (
	"io"
	"time"

	uuid "github.com/satori/go.uuid"
)

type FileItf interface {
	GetID() uuid.UUID
	GetName() string
	GetDescription() string
}

type File struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	Name        string
	Description string
}

func (f File) GetID() uuid.UUID {
	return f.ID
}

func (f File) GetName() string {
	return f.Name
}

func (f File) GetDescription() string {
	return f.Description
}

type Image struct {
	File
	Format   string
	Width    int
	Height   int
	ExifJSON string
}

type ImageFile interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}
