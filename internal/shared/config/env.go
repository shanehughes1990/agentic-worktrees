package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
)

var defaultValidator = validator.New(validator.WithRequiredStructEnabled())

func Decode(prefix string, target any) error {
	if target == nil {
		return fmt.Errorf("config target cannot be nil")
	}

	if err := envconfig.Process(prefix, target); err != nil {
		return err
	}

	return defaultValidator.Struct(target)
}

func Load[T any](prefix string) (T, error) {
	var cfg T
	if err := Decode(prefix, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
