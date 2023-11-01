package scalar_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/nais/console-backend/internal/graph/scalar"
	"github.com/stretchr/testify/assert"
)

var tm = time.Date(2020, time.April, 20, 0, 0, 0, 0, time.UTC)

func TestDate_NewDate(t *testing.T) {
	date := scalar.NewDate(tm)
	assert.Equal(t, "2020-04-20", date.String())
}

func TestDate_PgDate(t *testing.T) {
	date := scalar.NewDate(tm)
	assert.Equal(t, "2020-04-20", date.PgDate().Time.Format(scalar.DateFormatYYYYMMDD))
}

func TestDate_MarshalGQLContext(t *testing.T) {
	date := scalar.NewDate(tm)
	buf := new(bytes.Buffer)
	err := date.MarshalGQLContext(context.Background(), buf)
	assert.NoError(t, err)
	assert.Equal(t, `"2020-04-20"`, buf.String())
}

func TestDate_UnmarshalGQLContext(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid type", func(t *testing.T) {
		date := scalar.NewDate(tm)
		err := date.UnmarshalGQLContext(ctx, 123)
		assert.EqualError(t, err, "date must be a string")
	})

	t.Run("empty string", func(t *testing.T) {
		date := scalar.NewDate(tm)
		err := date.UnmarshalGQLContext(ctx, "")
		assert.EqualError(t, err, "date must not be empty")
	})

	t.Run("valid", func(t *testing.T) {
		date := scalar.NewDate(time.Now())
		err := date.UnmarshalGQLContext(ctx, "2020-04-20")
		assert.NoError(t, err)
		assert.Equal(t, "2020-04-20", string(date))
	})
}
