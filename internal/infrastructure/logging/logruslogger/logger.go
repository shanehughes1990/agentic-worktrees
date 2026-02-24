package logruslogger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func NewFromEnv() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	level := logrus.InfoLevel
	if rawLevel := strings.TrimSpace(os.Getenv("LOG_LEVEL")); rawLevel != "" {
		if parsedLevel, err := logrus.ParseLevel(rawLevel); err == nil {
			level = parsedLevel
		}
	}
	logger.SetLevel(level)

	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	return logger
}
