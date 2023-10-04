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
	CostLastDate(ctx context.Context) (pgtype.Date, error)
	CostUpsert(ctx context.Context, arg []CostUpsertParams) *CostUpsertBatchResults
	GetCost(ctx context.Context) ([]*Cost, error)
	MonthlyCostForApp(ctx context.Context, arg MonthlyCostForAppParams) ([]*MonthlyCostForAppRow, error)
	MonthlyCostForTeam(ctx context.Context, team *string) ([]*MonthlyCostForTeamRow, error)
}

var _ Querier = (*Queries)(nil)
