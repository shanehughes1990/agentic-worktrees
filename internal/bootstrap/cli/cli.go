package cli

import "context"

type Runtime struct{}

func New() (*Runtime, error) {
	return &Runtime{}, nil
}

func (r *Runtime) Run(ctx context.Context) error {
	_ = ctx
	return nil
}
