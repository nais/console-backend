package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nais/console-backend/internal/database"
)

func TestGetInstanceConnectionNameFromDsn(t *testing.T) {
	t.Run("empty dsn", func(t *testing.T) {
		instance, err := database.GetInstanceConnectionNameFromDsn("")
		assert.Equal(t, "", instance)
		assert.Equal(t, "dsn does not have a host field: \"\"", err.Error())
	})

	t.Run("valid dsn", func(t *testing.T) {
		instance, err := database.GetInstanceConnectionNameFromDsn("host=some-project-123:europe-north1:foo-bar-123 user=user@project-id-123.iam dbname=db_name sslmode=disable")
		assert.Equal(t, "some-project-123:europe-north1:foo-bar-123", instance)
		assert.NoError(t, err)
	})
}
