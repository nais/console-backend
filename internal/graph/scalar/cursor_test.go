package scalar_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/nais/console-backend/internal/graph/scalar"
	"github.com/stretchr/testify/assert"
)

func TestCursor_MarshalGQLContext(t *testing.T) {
	ctx := context.Background()

	t.Run("no offset", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cursor := scalar.Cursor{}
		err := cursor.MarshalGQLContext(ctx, buf)
		assert.NoError(t, err)
		assert.Equal(t, `"b2Zmc2V0PTA="`, buf.String())
	})

	t.Run("with offset", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cursor := scalar.Cursor{Offset: 42}
		err := cursor.MarshalGQLContext(ctx, buf)
		assert.NoError(t, err)
		assert.Equal(t, `"b2Zmc2V0PTQy"`, buf.String())
	})
}

func TestCursor_UnmarshalGQLContext(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid type", func(t *testing.T) {
		cursor := scalar.Cursor{}
		err := cursor.UnmarshalGQLContext(ctx, 123)
		assert.EqualError(t, err, "cursor must be a string")
	})

	t.Run("not base64", func(t *testing.T) {
		cursor := scalar.Cursor{}
		err := cursor.UnmarshalGQLContext(ctx, "foobar")
		assert.ErrorContains(t, err, "illegal base64")
	})

	t.Run("valid", func(t *testing.T) {
		cursor := scalar.Cursor{}
		err := cursor.UnmarshalGQLContext(ctx, "b2Zmc2V0PTQy")
		assert.NoError(t, err)
		assert.Equal(t, 42, cursor.Offset)
	})

	t.Run("invalid offset", func(t *testing.T) {
		cursor := scalar.Cursor{}
		err := cursor.UnmarshalGQLContext(ctx, "b2Zmc2V0PXd1dA==")
		assert.ErrorContains(t, err, "invalid syntax")
	})
}
