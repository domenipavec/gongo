package files

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/matematik7/gongo"
	"github.com/matematik7/gongo/files/storage"
	"github.com/matematik7/gongo/render"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/xor-gate/goexif2/exif"
)

type Files struct {
	storage storage.Storage

	db *gorm.DB
}

func New(storage storage.Storage) *Files {
	return &Files{
		storage: storage,
	}
}

func (f *Files) Configure(app gongo.App) error {
	f.db = app["DB"].(*gorm.DB)

	app["Render"].(*render.Render).AddContextFunc(func(r *http.Request, ctx render.Context) {
		ctx["file_url"] = func(file FileItf) string {
			url, err := f.URL(file)
			if err != nil {
				// TODO: handle this error
			}
			return url
		}
	})

	return nil
}

func (f *Files) Resources() []interface{} {
	return []interface{}{
		&File{},
		&Image{},
	}
}

func (f *Files) newFile(input io.Reader, name, description string) (File, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return File{}, errors.Wrap(err, "could not generate uuid for file")
	}

	file := File{
		ID:          uuid,
		Name:        name,
		Description: description,
	}

	err = f.storage.Save(file.ID.String(), input)
	if err != nil {
		return file, errors.Wrap(err, "could not save file to storage")
	}

	return file, nil
}

func (f *Files) NewFile(input io.Reader, name, description string) (File, error) {
	file, err := f.newFile(input, name, description)
	if err != nil {
		return file, err
	}

	if err := f.db.Create(&file).Error; err != nil {
		return file, errors.Wrap(err, "could not save file to db")
	}

	return file, nil
}

func (f *Files) NewImage(input ImageFile, name, description string) (Image, error) {
	img := Image{}

	cfg, format, err := image.DecodeConfig(input)
	if err != nil {
		return img, errors.Wrap(err, "could not decode image format")
	}
	img.Format = format
	img.Width = cfg.Width
	img.Height = cfg.Height

	if _, err := input.Seek(0, io.SeekStart); err != nil {
		return img, errors.Wrap(err, "could not seek to start")
	}

	if x, err := exif.Decode(input); err == nil {
		exifData, err := x.MarshalJSON()
		if err != nil {
			return img, errors.Wrap(err, "could not marshal exif data")
		}
		img.ExifJSON = string(exifData)
	}

	if _, err := input.Seek(0, io.SeekStart); err != nil {
		return img, errors.Wrap(err, "could not seek to start")
	}

	file, err := f.newFile(input, name, description)
	if err != nil {
		return img, err
	}
	img.File = file

	if err := f.db.Create(&img).Error; err != nil {
		return img, errors.Wrap(err, "could not save image to db")
	}

	return img, nil
}

func (f *Files) Delete(file FileItf) error {
	tx := f.db.Begin()

	if err := tx.Delete(file).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "could not delete file from db")
	}

	if err := f.storage.Delete(file.GetID().String()); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "could not delete file from storage")
	}

	tx.Commit()

	return nil
}

func (f *Files) URL(file FileItf) (string, error) {
	return f.storage.URL(file.GetID().String())
}
