package graph

import (
	"fmt"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/hookd"
	"github.com/nais/console-backend/internal/k8s"
	"github.com/nais/console-backend/internal/search"
	"github.com/nais/console-backend/internal/teams"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/metric"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	hookdClient *hookd.Client
	teamsClient *teams.Client
	k8sClient   *k8s.Client
	searcher    *search.Searcher
	log         logrus.FieldLogger
	querier     gensql.Querier
	clusters    []string
}

// NewHandler creates and returns a new GraphQL handler with the given schema
func NewHandler(hookdClient *hookd.Client, teamsClient *teams.Client, k8sClient *k8s.Client, querier gensql.Querier, clusters []string, log logrus.FieldLogger, meter metric.Meter) (*handler.Server, error) {
	metricsMiddleware, err := NewMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("create metrics middleware: %w", err)
	}

	config := Config{
		Resolvers: &Resolver{
			hookdClient: hookdClient,
			teamsClient: teamsClient,
			k8sClient:   k8sClient,
			searcher:    search.New(teamsClient, k8sClient),
			log:         log,
			querier:     querier,
			clusters:    clusters,
		},
	}
	schema := NewExecutableSchema(config)
	graphHandler := handler.New(schema)
	graphHandler.Use(metricsMiddleware)
	graphHandler.AddTransport(transport.SSE{}) // Support subscriptions
	graphHandler.AddTransport(transport.Options{})
	graphHandler.AddTransport(transport.POST{})
	graphHandler.SetQueryCache(lru.New(1000))
	graphHandler.Use(extension.Introspection{})
	graphHandler.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})
	return graphHandler, nil
}
