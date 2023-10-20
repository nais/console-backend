package logger

import (
	"fmt"
	"strings"

	"github.com/nais/console-backend/internal/config"
	"github.com/sirupsen/logrus"
)

// New creates a new logger with the given format and level
func New(cfg config.Logger) (logrus.FieldLogger, error) {
	log := logrus.StandardLogger()

	switch strings.ToLower(cfg.Format) {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		log.SetFormatter(&logrus.TextFormatter{})
	default:
		return nil, fmt.Errorf("invalid log format: %q", cfg.Format)
	}

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	log.SetLevel(level)
	return log, nil
}
