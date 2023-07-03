package k8s_test

import (
	"strings"
	"testing"

	"github.com/nais/console-backend/internal/k8s"
	"github.com/sirupsen/logrus"
)

func TestNew(t *testing.T) {
	t.Run("Fails with invalid static cluster entry", func(t *testing.T) {
		_, err := k8s.New([]string{"cluster"}, []string{"invalid"}, "", "", nil, nil)
		if err == nil {
			t.Fatal("Should fail with invalid static cluster entry")
		}
		want := "invalid static cluster entry: invalid. Must be on format 'name|apiserver-host|token'"
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("got: %s, want: %s", err.Error(), want)
		}
	})

	t.Run("Happy path", func(t *testing.T) {
		_, err := k8s.New([]string{"cluster"}, nil, "tenant", "", nil, logrus.NewEntry(logrus.New()))
		if err != nil {
			t.Fatalf("Should not fail: %s", err.Error())
		}
	})

	t.Run("Happy path with static clusters", func(t *testing.T) {
		_, err := k8s.New([]string{"cluster"}, []string{"cluster|host|token"}, "tenant", "", nil, logrus.NewEntry(logrus.New()))
		if err != nil {
			t.Fatalf("Should not fail with valid static cluster entry: %s", err.Error())
		}
	})
}
