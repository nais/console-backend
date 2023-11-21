package dtrack

import dependencytrack "github.com/nais/dependencytrack/pkg/client"

type DependencytrackClient interface {
	dependencytrack.Client
}
