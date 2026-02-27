// Package observability provides logging, tracing, metrics, and correlation primitives
// for the V1 control plane.
package observability

import (
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// DefaultTimestampFormat is the shared timestamp layout used by text and JSON log output.
	DefaultTimestampFormat = time.RFC3339Nano
)

// LoggerType selects the output formatter used by logrus.
type LoggerType string

const (
	// LoggerTypeJSON configures logrus JSON output.
	LoggerTypeJSON LoggerType = "json"
	// LoggerTypeText configures logrus text output.
	LoggerTypeText LoggerType = "text"
)

// LoggerOptions configures logrus output type, level, and formatter settings.
type LoggerOptions struct {
	Type            LoggerType
	Level           logrus.Level
	TimestampFormat string
	PrettyPrintJSON bool
}

// DefaultLoggerOptions returns the baseline logger configuration for V1.
//
// By default the logger uses text output with RFC3339Nano timestamps at info level.
func DefaultLoggerOptions() LoggerOptions {
	return LoggerOptions{
		Type:            LoggerTypeText,
		Level:           logrus.InfoLevel,
		TimestampFormat: DefaultTimestampFormat,
	}
}

func (options LoggerOptions) withDefaults() LoggerOptions {
	defaults := DefaultLoggerOptions()
	if options.Type == "" {
		options.Type = defaults.Type
	}
	if options.TimestampFormat == "" {
		options.TimestampFormat = defaults.TimestampFormat
	}
	if options.Level == 0 {
		options.Level = defaults.Level
	}
	return options
}

// NewLogrusLogger builds a configured logrus logger using the provided options.
//
// JSON output is enabled only when options.Type is LoggerTypeJSON. Both JSON and
// text output use the same timestamp format to keep logs comparable across formats.
func NewLogrusLogger(options LoggerOptions) *logrus.Logger {
	options = options.withDefaults()

	logger := logrus.New()
	logger.SetLevel(options.Level)
	switch options.Type {
	case LoggerTypeJSON:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat:   options.TimestampFormat,
			DisableHTMLEscape: true,
			PrettyPrint:       options.PrettyPrintJSON,
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        options.TimestampFormat,
			DisableLevelTruncation: true,
			PadLevelText:           true,
			DisableSorting:         false,
			ForceQuote:             true,
			QuoteEmptyFields:       true,
			DisableQuote:           false,
		})
	}

	return logger
}
