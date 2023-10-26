package model_test

import (
	"testing"

	"github.com/nais/console-backend/internal/graph/scalar"

	"github.com/nais/console-backend/internal/graph/model"
	"github.com/stretchr/testify/assert"
)

func Test_NewPagination(t *testing.T) {
	t.Run("no pagination", func(t *testing.T) {
		pagination, err := model.NewPagination(nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, 10, pagination.First())
		assert.Equal(t, 10, pagination.Last())
		assert.Equal(t, -1, pagination.After().Offset)
		assert.Nil(t, pagination.Before())
	})

	t.Run("both first and last should fail", func(t *testing.T) {
		pagination, err := model.NewPagination(intP(10), intP(10), nil, nil)
		assert.Nil(t, pagination)
		assert.EqualError(t, err, "using both `first` and `last` with pagination is not supported")
	})

	t.Run("pagination with values", func(t *testing.T) {
		pagination, err := model.NewPagination(intP(9), nil, &scalar.Cursor{Offset: 8}, &scalar.Cursor{Offset: 7})
		assert.NoError(t, err)
		assert.Equal(t, 9, pagination.First())
		assert.Equal(t, 10, pagination.Last())
		assert.Equal(t, 8, pagination.After().Offset)
		assert.Equal(t, 7, pagination.Before().Offset)
	})
}

func intP(i int) *int {
	return &i
}
