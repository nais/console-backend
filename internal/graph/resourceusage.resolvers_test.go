package graph_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nais/console-backend/internal/graph"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
	"github.com/nais/console-backend/internal/resourceusage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_queryResolver_ResourceUtilizationForApp(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid from date", func(t *testing.T) {
		from := "from"
		fromDate := scalar.Date(from)
		resp, err := graph.
			NewResolver(nil, nil, nil, nil, nil, nil, nil).
			Query().
			ResourceUtilizationForApp(ctx, "env", "team", "app", &fromDate, nil)
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, fmt.Sprintf("cannot parse %q", from))
	})

	t.Run("invalid to date", func(t *testing.T) {
		to := "to"
		toDate := scalar.Date(to)
		resp, err := graph.
			NewResolver(nil, nil, nil, nil, nil, nil, nil).
			Query().
			ResourceUtilizationForApp(ctx, "env", "team", "app", nil, &toDate)
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, fmt.Sprintf("cannot parse %q", to))
	})

	t.Run("no dates specified", func(t *testing.T) {
		cpuData := make([]model.ResourceUtilization, 0)
		memoryData := make([]model.ResourceUtilization, 0)

		resourceUsageClient := resourceusage.NewMockClient(t)
		resourceUsageClient.
			EXPECT().
			ResourceUtilizationForApp(ctx, "env", "team", "app", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Run(func(_ context.Context, _, _, _ string, start time.Time, end time.Time) {
				allowedDelta := time.Second
				assert.WithinDuration(t, end, time.Now(), allowedDelta)
				assert.WithinDuration(t, start, time.Now(), 24*6*time.Hour+allowedDelta)
			}).
			Return(&model.ResourceUtilizationForApp{
				CPU:    cpuData,
				Memory: memoryData,
			}, nil)

		resp, err := graph.
			NewResolver(nil, nil, nil, resourceUsageClient, nil, nil, nil).
			Query().
			ResourceUtilizationForApp(ctx, "env", "team", "app", nil, nil)
		assert.Equal(t, cpuData, resp.CPU)
		assert.Equal(t, memoryData, resp.Memory)
		assert.NoError(t, err)
	})

	t.Run("both dates specified", func(t *testing.T) {
		fromTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		from := scalar.NewDate(fromTime)
		toTime := time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)
		to := scalar.NewDate(toTime)

		cpuData := make([]model.ResourceUtilization, 0)
		memoryData := make([]model.ResourceUtilization, 0)

		resourceUsageClient := resourceusage.NewMockClient(t)
		resourceUsageClient.
			EXPECT().
			ResourceUtilizationForApp(ctx, "env", "team", "app", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Run(func(_ context.Context, _, _, _ string, start time.Time, end time.Time) {
				assert.Equal(t, fromTime, start)
				assert.Equal(t, toTime, end)
			}).
			Return(&model.ResourceUtilizationForApp{
				CPU:    cpuData,
				Memory: memoryData,
			}, nil)

		resp, err := graph.
			NewResolver(nil, nil, nil, resourceUsageClient, nil, nil, nil).
			Query().
			ResourceUtilizationForApp(ctx, "env", "team", "app", &from, &to)
		assert.NoError(t, err)
		assert.Equal(t, cpuData, resp.CPU)
		assert.Equal(t, memoryData, resp.Memory)
	})
}