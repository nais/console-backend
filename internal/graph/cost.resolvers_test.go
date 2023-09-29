package graph_test

import (
	"context"
	"testing"

	"github.com/nais/console-backend/internal/graph"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/stretchr/testify/assert"
)

func Test_costResolver_Sum(t *testing.T) {
	ctx := context.Background()
	resolver := &graph.Resolver{}
	costResolver := resolver.Cost()

	t.Run("empty cost", func(t *testing.T) {
		sum, err := costResolver.Sum(ctx, &model.Cost{})
		assert.NoError(t, err)
		assert.Equal(t, 0.0, sum)
		assert.InDelta(t, 0.0, sum, 0.0001)
	})

	t.Run("non-empty cost", func(t *testing.T) {
		sum, err := costResolver.Sum(ctx, &model.Cost{
			Series: []*model.CostSeries{
				{
					Data: []*model.DailyCost{
						{Cost: 1.0},
						{Cost: 3.0},
						{Cost: 5.5},
					},
				},
				{
					Data: []*model.DailyCost{
						{Cost: 2.0},
						{Cost: 4.0},
					},
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 15.5, sum)
	})
}

func Test_costSeriesResolver_Sum(t *testing.T) {
	ctx := context.Background()
	resolver := &graph.Resolver{}
	costSeriesResolver := resolver.CostSeries()

	t.Run("empty cost", func(t *testing.T) {
		sum, err := costSeriesResolver.Sum(ctx, &model.CostSeries{})
		assert.NoError(t, err)
		assert.Equal(t, 0.0, sum)
		assert.InDelta(t, 0.0, sum, 0.0001)
	})

	t.Run("non-empty cost", func(t *testing.T) {
		sum, err := costSeriesResolver.Sum(ctx, &model.CostSeries{
			Data: []*model.DailyCost{
				{Cost: 1.0},
				{Cost: 2.0},
				{Cost: 3.0},
				{Cost: 4.0},
				{Cost: 5.5},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 15.5, sum)
	})
}
