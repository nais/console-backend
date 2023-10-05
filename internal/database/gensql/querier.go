// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0

package gensql

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	CostForApp(ctx context.Context, arg CostForAppParams) ([]*Cost, error)
	CostUpsert(ctx context.Context, arg []CostUpsertParams) *CostUpsertBatchResults
	// DailyCostForTeam will fetch the daily cost for a specific team across all apps and envs in a date range.
	DailyCostForTeam(ctx context.Context, arg DailyCostForTeamParams) ([]*Cost, error)
	EnvCostForTeam(ctx context.Context, arg EnvCostForTeamParams) ([]*EnvCostForTeamRow, error)
	// LastCostDate will return the last date that has a cost.
	LastCostDate(ctx context.Context) (pgtype.Date, error)
	MonthlyCostForApp(ctx context.Context, arg MonthlyCostForAppParams) ([]*MonthlyCostForAppRow, error)
	MonthlyCostForTeam(ctx context.Context, team *string) ([]*MonthlyCostForTeamRow, error)
}

var _ Querier = (*Queries)(nil)
