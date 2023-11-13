package model

import (
	"strings"

	"github.com/nais/console-backend/internal/database/gensql"
)

// ToDatabaseEnum converts a ResourceType to the database enum type
func (t ResourceType) ToDatabaseEnum() gensql.ResourceType {
	return gensql.ResourceType(strings.ToLower(string(t)))
}
