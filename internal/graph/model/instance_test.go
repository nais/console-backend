package model_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/nais/console-backend/internal/graph/model"
	"github.com/stretchr/testify/assert"
)

func TestInstanceState(t *testing.T) {
	t.Run("invalid state", func(t *testing.T) {
		state := model.InstanceState("foobar")
		assert.False(t, state.IsValid())
	})

	t.Run("valid state", func(t *testing.T) {
		state := model.InstanceStateFailing
		assert.True(t, state.IsValid())
		assert.Equal(t, "FAILING", state.String())
	})
}

func TestInstanceState_MarshalGQLContext(t *testing.T) {
	ctx := context.Background()
	buf := new(bytes.Buffer)
	state := model.InstanceStateRunning
	err := state.MarshalGQLContext(ctx, buf)
	assert.NoError(t, err)
	assert.Equal(t, `"RUNNING"`, buf.String())
}

func TestInstanceState_UnmarshalGQLContext(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid type", func(t *testing.T) {
		state := model.InstanceStateRunning
		err := state.UnmarshalGQLContext(ctx, 123)
		assert.EqualError(t, err, "instance state must be a string")
	})

	t.Run("invalid", func(t *testing.T) {
		state := model.InstanceStateRunning
		err := state.UnmarshalGQLContext(ctx, "foobar")
		assert.ErrorContains(t, err, `"foobar" is not a valid InstanceState`)
	})

	t.Run("valid", func(t *testing.T) {
		state := model.InstanceStateRunning
		err := state.UnmarshalGQLContext(ctx, string(model.InstanceStateFailing))
		assert.NoError(t, err)
		assert.Equal(t, model.InstanceStateFailing, state)
	})
}
