// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0

package gensql

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	// CostUpsert will insert or update a cost record. If there is a conflict on the daily_cost_key constrant, the
	// daily_cost column will be updated.
	CostUpsert(ctx context.Context, arg []CostUpsertParams) *CostUpsertBatchResults
	// DailyCostForApp will fetch the daily cost for a specific team app in a specific environment, across all cost types
	// in a date range.
	DailyCostForApp(ctx context.Context, arg DailyCostForAppParams) ([]*Cost, error)
	// DailyCostForTeam will fetch the daily cost for a specific team across all apps and envs in a date range.
	DailyCostForTeam(ctx context.Context, arg DailyCostForTeamParams) ([]*Cost, error)
	// DailyEnvCostForTeam will fetch the daily cost for a specific team and env across all apps in a date range.
	DailyEnvCostForTeam(ctx context.Context, arg DailyEnvCostForTeamParams) ([]*DailyEnvCostForTeamRow, error)
	// LastCostDate will return the last date that has a cost.
	LastCostDate(ctx context.Context) (pgtype.Date, error)
	// MaxResourceUtilizationDate will return the max date for resource utilization records.
	MaxResourceUtilizationDate(ctx context.Context) (pgtype.Timestamptz, error)
	MonthlyCostForApp(ctx context.Context, arg MonthlyCostForAppParams) ([]*MonthlyCostForAppRow, error)
	MonthlyCostForTeam(ctx context.Context, team *string) ([]*MonthlyCostForTeamRow, error)
	// ResourceUtilizationForApp will return resource utilization records for a given app.
	ResourceUtilizationForApp(ctx context.Context, arg ResourceUtilizationForAppParams) ([]*ResourceUtilizationMetric, error)
	// ResourceUtilizationForTeam will return resource utilization records for a given team.
	ResourceUtilizationForTeam(ctx context.Context, arg ResourceUtilizationForTeamParams) ([]*ResourceUtilizationForTeamRow, error)
	// ResourceUtilizationOverageCostForTeam will return overage records for a given team.
	ResourceUtilizationOverageCostForTeam(ctx context.Context, arg ResourceUtilizationOverageCostForTeamParams) ([]*ResourceUtilizationOverageCostForTeamRow, error)
	// ResourceUtilizationRangeForApp will return the min and max timestamps for a specific app.
	ResourceUtilizationRangeForApp(ctx context.Context, arg ResourceUtilizationRangeForAppParams) (*ResourceUtilizationRangeForAppRow, error)
	// ResourceUtilizationRangeForTeam will return the min and max timestamps for a specific team.
	ResourceUtilizationRangeForTeam(ctx context.Context, team string) (*ResourceUtilizationRangeForTeamRow, error)
	// ResourceUtilizationUpsert will insert or update resource utilization records.
	ResourceUtilizationUpsert(ctx context.Context, arg []ResourceUtilizationUpsertParams) *ResourceUtilizationUpsertBatchResults
	// TruncateCostTable will truncate the cost table before doing a complete reimport.
	TruncateCostTable(ctx context.Context) error
}

var _ Querier = (*Queries)(nil)
