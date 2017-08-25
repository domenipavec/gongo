package gongo

import "context"

type CallbackFunc func(ctx context.Context) error

type Callback struct {
	callbacks []CallbackFunc
}

func (c *Callback) Add(f CallbackFunc) {
	c.callbacks = append(c.callbacks, f)
}

func (c Callback) Call(ctx context.Context) error {
	for _, cf := range c.callbacks {
		if err := cf(ctx); err != nil {
			return err
		}
	}

	return nil
}
