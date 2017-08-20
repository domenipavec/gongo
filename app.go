package gongo

import (
	"github.com/pkg/errors"
)

type Configurer interface {
	Configure(app App) error
}

type Resourcer interface {
	Resources() []interface{}
}

type App map[string]interface{}

func (app App) Configure() error {
	for name, itf := range app {
		if configurer, ok := itf.(Configurer); ok {
			if err := configurer.Configure(app); err != nil {
				return errors.Wrapf(err, "could not configure %s", name)
			}
		}
	}

	return nil
}
