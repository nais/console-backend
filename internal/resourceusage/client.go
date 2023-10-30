package resourceusage

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prom "github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

type clusterName string

type Client interface {
	Query(ctx context.Context, env, team, app string) (prom.Value, error)
}

type client struct {
	promClients map[clusterName]promv1.API
	log         logrus.FieldLogger
}

func New(clusters []string, tenant string, log logrus.FieldLogger) (Client, error) {
	promClients := map[clusterName]promv1.API{}
	for _, cluster := range clusters {
		promClient, err := api.NewClient(api.Config{
			Address: fmt.Sprintf("https://prometheus.%s.%s.cloud.nais.io", cluster, tenant),
		})
		if err != nil {
			return nil, err
		}
		promClients[clusterName(cluster)] = promv1.NewAPI(promClient)
	}

	return &client{
		promClients: promClients,
		log:         log,
	}, nil
}

func (c *client) Query(ctx context.Context, env, team, app string) (prom.Value, error) {
	promClient, exists := c.promClients[clusterName(env)]
	if !exists {
		return nil, fmt.Errorf("no prometheus client for cluster: %q", env)
	}

	query := fmt.Sprintf(
		`avg(rate(container_cpu_usage_seconds_total{namespace=%q, cpu="total", container=%q}[1h])) by (container, namespace)`,
		team,
		app,
	)
	to := time.Now()
	from := to.Add(-24 * time.Hour)
	value, _, err := promClient.QueryRange(ctx, query, promv1.Range{
		Start: from,
		End:   to,
		Step:  time.Hour,
	})
	return value, err
}
