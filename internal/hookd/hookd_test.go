package hookd

import (
	"context"
	"testing"
)

func TestGetDeploysForTeam(t *testing.T) {
	c := New("secret-frontend-psk", "http://hookd.local.nais.io", nil, nil)
	teamName := "b"
	deploys, err := c.Deployments(context.Background(), WithTeam(teamName))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("deploys: %v", deploys)
}
