package teams_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/teams"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

func TestClient_Search(t *testing.T) {
	const apiToken = "token"
	ctx := context.Background()
	testLogger, _ := test.NewNullLogger()
	log := testLogger.WithContext(ctx)

	t.Run("no teams filter", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{})
		searchType := model.SearchTypeApp
		results := teams.
			New(apiToken, teamsBackend.URL, errorsMeter(t), log).
			Search(ctx, "query", &model.SearchFilter{Type: &searchType})
		assert.Nil(t, results)
	})

	t.Run("no teams from the teams-backend", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "Bearer "+apiToken, r.Header.Get("Authorization"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{}"))
			},
		})
		searchType := model.SearchTypeTeam
		results := teams.
			New(apiToken, teamsBackend.URL, errorsMeter(t), log).
			Search(ctx, "query", &model.SearchFilter{Type: &searchType})
		assert.NotNil(t, results)
		assert.Empty(t, results)
	})

	t.Run("teams from the teams-backend but no match", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": {"teams": [{"slug": "team-1"}, {"slug": "team-2"}]}}`))
			},
		})
		searchType := model.SearchTypeTeam
		results := teams.
			New(apiToken, teamsBackend.URL, errorsMeter(t), log).
			Search(ctx, "query", &model.SearchFilter{Type: &searchType})
		assert.NotNil(t, results)
		assert.Empty(t, results)
	})

	t.Run("teams from the teams-backend with matches", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": {"teams": [{"slug": "foo-team"}, {"slug": "team-foo"}, {"slug": "team-foo-bar"}, {"slug": "team"}]}}`))
			},
		})
		searchType := model.SearchTypeTeam
		results := teams.
			New(apiToken, teamsBackend.URL, errorsMeter(t), log).
			Search(ctx, "foo", &model.SearchFilter{Type: &searchType})
		assert.Len(t, results, 3)

		team1, _ := results[0].Node.(*model.Team)
		team2, _ := results[1].Node.(*model.Team)
		team3, _ := results[2].Node.(*model.Team)

		assert.Equal(t, "foo-team", team1.Name)
		assert.Equal(t, "team-foo", team2.Name)
		assert.Equal(t, "team-foo-bar", team3.Name)
	})
}

func httpServerWithHandlers(t *testing.T, handlers []http.HandlerFunc) *httptest.Server {
	idx := 0
	t.Cleanup(func() {
		diff := len(handlers) - idx
		if diff != 0 {
			t.Fatalf("too many configured handlers, remove %d handler(s)", diff)
		}
	})
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(handlers) < idx+1 {
			t.Fatalf("unexpected request, add missing handler func: %v", r)
		}
		handlers[idx](w, r)
		idx += 1
	}))
}

func errorsMeter(t *testing.T) api.Int64Counter {
	t.Helper()

	meter := metric.NewMeterProvider().Meter("github.com/nais/console-backend")
	errors, _ := meter.Int64Counter("errors")
	return errors
}
