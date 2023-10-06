// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0

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
	MonthlyCostForApp(ctx context.Context, arg MonthlyCostForAppParams) ([]*MonthlyCostForAppRow, error)
	MonthlyCostForTeam(ctx context.Context, team *string) ([]*MonthlyCostForTeamRow, error)
}

var _ Querier = (*Queries)(nil)
