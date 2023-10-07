package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nais/console-backend/internal/database"
)

func TestExtractInstanceConnectionNameFromDsn(t *testing.T) {
	t.Run("empty dsn", func(t *testing.T) {
		dsn, instance, err := database.ExtractInstanceConnectionNameFromDsn("")
		assert.Equal(t, "", dsn)
		assert.Equal(t, "", instance)
		assert.Equal(t, "dsn does not have a host field: \"\"", err.Error())
	})

	t.Run("valid dsn", func(t *testing.T) {
		dsn, instance, err := database.ExtractInstanceConnectionNameFromDsn("host=some-project-123:europe-north1:foo-bar-123 user=user@project-id-123.iam dbname=db_name sslmode=disable")
		assert.Equal(t, "dbname=db_name sslmode=disable user=user@project-id-123.iam", dsn)
		assert.Equal(t, "some-project-123:europe-north1:foo-bar-123", instance)
		assert.NoError(t, err)
	})
}
