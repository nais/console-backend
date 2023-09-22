package graph

import (
	"github.com/nais/console-backend/internal/hookd"
	"github.com/nais/console-backend/internal/k8s"
	"github.com/nais/console-backend/internal/search"
	"github.com/nais/console-backend/internal/teams"
	"github.com/nais/console-backend/pkg/database"
	"github.com/sirupsen/logrus"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Hookd       *hookd.Client
	TeamsClient *teams.Client
	K8s         *k8s.Client
	Searcher    *search.Searcher
	Log         *logrus.Logger
	Repo        *database.Repo
}
