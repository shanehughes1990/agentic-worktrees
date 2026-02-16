package app

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	sharedconfig "github.com/shanehughes1990/agentic-worktrees/internal/shared/config"
	"github.com/shanehughes1990/agentic-worktrees/internal/shared/logging"
)

type Kind string

const (
	KindCLI    Kind = "cli"
	KindWorker Kind = "worker"
)

type Config struct {
	Name      string `envconfig:"NAME" default:"agentic-worktrees" validate:"required"`
	LogLevel  string `envconfig:"LOG_LEVEL" default:"info" validate:"required,oneof=trace debug info warn warning error fatal panic"`
	LogFormat string `envconfig:"LOG_FORMAT" default:"json" validate:"required,oneof=json text"`
}

type Runtime struct {
	kind   Kind
	config Config
	logger *logrus.Logger
}

func Init(kind Kind) (*Runtime, error) {
	if kind != KindCLI && kind != KindWorker {
		return nil, fmt.Errorf("unsupported runtime kind %q", kind)
	}

	cfg, err := sharedconfig.Load[Config]("APP")
	if err != nil {
		return nil, fmt.Errorf("load app config: %w", err)
	}

	logger, err := logging.New(logging.Config{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
	})
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	return &Runtime{kind: kind, config: cfg, logger: logger}, nil
}

func Run(ctx context.Context, runtime *Runtime) error {
	if runtime == nil {
		return fmt.Errorf("runtime cannot be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	switch runtime.kind {
	case KindCLI:
		return runtime.runCLI(ctx)
	case KindWorker:
		return runtime.runWorker(ctx)
	default:
		return fmt.Errorf("unsupported runtime kind %q", runtime.kind)
	}
}
