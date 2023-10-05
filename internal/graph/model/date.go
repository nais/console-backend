package model

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const DateFormat = "2006-01-02"

type Date string

func (d Date) MarshalGQLContext(_ context.Context, w io.Writer) error {
	_, err := io.WriteString(w, strconv.Quote(string(d)))
	if err != nil {
		return fmt.Errorf("writing date: %w", err)
	}
	return nil
}

func (d *Date) UnmarshalGQLContext(_ context.Context, v interface{}) error {
	date, ok := v.(string)
	if !ok {
		return fmt.Errorf("date must be a string")
	}

	if date == "" {
		return fmt.Errorf("date must not be empty")
	}

	*d = Date(date)
	return nil
}

// NewDate returns a Date from a time.Time
func NewDate(t time.Time) Date {
	return Date(t.UTC().Format(DateFormat))
}

// String returns the Date as a string
func (d Date) String() string {
	return string(d)
}

// PgDate returns the Date as a pgtype.Date instance
func (d Date) PgDate() (date pgtype.Date) {
	date.Scan(string(d))
	return date
}
