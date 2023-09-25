package database

import (
	"context"
	"time"
)

type CostRepo interface {
	CostLastDate(ctx context.Context) (time.Time, error)
}
