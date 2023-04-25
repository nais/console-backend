package graph

import (
	"github.com/nais/console-backend/internal/console"
	"github.com/nais/console-backend/internal/hookd"
	"github.com/nais/console-backend/internal/k8s"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Hookd   *hookd.Client
	Console *console.Client
	K8s     *k8s.Client
}
