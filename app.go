package gongo

import (
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type App struct {
	Authentication Authentication
	Authorization  Authorization
	DB             *gorm.DB
	Resources      Resources
	Store          sessions.Store
	Controllers    []Controller
}

func (app App) allControllers() []Controller {
	return append(
		[]Controller{
			app.Authentication,
			app.Authorization,
			app.Resources,
		},
		app.Controllers...,
	)
}

func (app App) Configure() error {
	for _, controller := range app.allControllers() {
		if err := controller.Configure(app); err != nil {
			return errors.Wrapf(err, "could not configure %s", controller.Name())
		}
	}

	return nil
}

func (app App) RegisterResources() error {
	if app.Resources == nil {
		return nil
	}

	for _, controller := range app.allControllers() {
		resources := controller.Resources()
		if resources != nil {
			if err := app.Resources.Register(controller.Name(), resources...); err != nil {
				return errors.Wrapf(err, "could not register resources for %s", controller.Name())
			}
		}
	}

	return nil
}
