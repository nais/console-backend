package resourceusage

import (
	"context"
	"fmt"
	"time"

	"github.com/nais/console-backend/internal/graph/model"

	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prom "github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

type clusterName string

type Client interface {
	DailyCpuUsageForApp(ctx context.Context, env, team, app string) ([]model.ResourceUsageValue, error)
	DailyMemoryUsageForApp(ctx context.Context, env, team, app string) ([]model.ResourceUsageValue, error)
	DailyCpuRequestForApp(ctx context.Context, env, team, app string) ([]model.ResourceUsageValue, error)
	DailyMemoryRequestForApp(ctx context.Context, env, team, app string) ([]model.ResourceUsageValue, error)
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

func (c *client) DailyCpuUsageForApp(ctx context.Context, env, team, app string) ([]model.ResourceUsageValue, error) {
	promClient, exists := c.promClients[clusterName(env)]
	if !exists {
		return nil, fmt.Errorf("no prometheus client for cluster: %q", env)
	}

	query := fmt.Sprintf(
		`sum(rate(container_cpu_usage_seconds_total{namespace=%q, container=%q}[1h]))`,
		team,
		app,
	)
	return rangedQuery(ctx, promClient, query)
}

func (c *client) DailyMemoryUsageForApp(ctx context.Context, env, team, app string) ([]model.ResourceUsageValue, error) {
	promClient, exists := c.promClients[clusterName(env)]
	if !exists {
		return nil, fmt.Errorf("no prometheus client for cluster: %q", env)
	}

	query := fmt.Sprintf(
		`max(max_over_time(container_memory_usage_bytes{namespace=%q, container=%q}[1h]))`,
		team,
		app,
	)
	return rangedQuery(ctx, promClient, query)
}

func (c *client) DailyCpuRequestForApp(ctx context.Context, env, team, app string) ([]model.ResourceUsageValue, error) {
	promClient, exists := c.promClients[clusterName(env)]
	if !exists {
		return nil, fmt.Errorf("no prometheus client for cluster: %q", env)
	}

	query := fmt.Sprintf(
		`max(max_over_time(kube_pod_container_resource_requests{container=%q, namespace=%q, resource="cpu", unit="core"}[5m]))`,
		app,
		team,
	)
	return rangedQuery(ctx, promClient, query)
}

func (c *client) DailyMemoryRequestForApp(ctx context.Context, env, team, app string) ([]model.ResourceUsageValue, error) {
	promClient, exists := c.promClients[clusterName(env)]
	if !exists {
		return nil, fmt.Errorf("no prometheus client for cluster: %q", env)
	}

	query := fmt.Sprintf(
		`max(max_over_time(kube_pod_container_resource_requests{container=%q, namespace=%q, resource="memory", unit="byte"}[5m]))`,
		app,
		team,
	)
	return rangedQuery(ctx, promClient, query)
}

func rangedQuery(ctx context.Context, client promv1.API, query string) ([]model.ResourceUsageValue, error) {
	to := time.Now()
	from := to.Add(-24 * time.Hour * 30)
	value, _, err := client.QueryRange(ctx, query, promv1.Range{
		Start: from,
		End:   to,
		Step:  time.Hour,
	})
	if err != nil {
		return nil, err
	}

	usage := make([]model.ResourceUsageValue, 0)
	if len(value.(prom.Matrix)) == 0 {
		return usage, nil
	}

	for _, val := range value.(prom.Matrix)[0].Values {
		usage = append(usage, model.ResourceUsageValue{
			Timestamp: val.Timestamp.Time(),
			Value:     float64(val.Value),
		})
	}

	return usage, nil
}
