// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0

package gensql

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Cost struct {
	ID       int32
	Env      *string
	Team     *string
	App      *string
	CostType string
	Date     pgtype.Date
	Cost     float32
}
