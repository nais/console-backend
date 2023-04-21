package hookd

import (
	"context"
	"testing"
)

func TestGetDeploysForTeam(t *testing.T) {
	c := New("secret-frontend-psk", "http://hookd.local.nais.io")
	deploys, err := c.GetDeploysForTeam(context.Background(), "b")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("deploys: %v", deploys)
}
