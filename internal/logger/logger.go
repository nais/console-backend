package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// New creates a new logger with the given format and level
func New(logFormat, logLevel string) (logrus.FieldLogger, error) {
	log := logrus.StandardLogger()

	switch logFormat {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		log.SetFormatter(&logrus.TextFormatter{})
	default:
		return nil, fmt.Errorf("invalid log format: %q", logFormat)
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}

	log.SetLevel(level)
	return log, nil
}
