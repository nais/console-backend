package teams_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/config"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/teams"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

const apiToken = "token"

func TestClient_Search(t *testing.T) {
	ctx := context.Background()
	testLogger, _ := test.NewNullLogger()
	log := testLogger.WithContext(ctx)

	t.Run("no teams filter", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{})
		searchType := model.SearchTypeApp
		results := teams.
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, false, errorsMeter(t), log).
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
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, false, errorsMeter(t), log).
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
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, false, errorsMeter(t), log).
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
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, false, errorsMeter(t), log).
			Search(ctx, "foo", &model.SearchFilter{Type: &searchType})
		assert.Len(t, results, 3)

		team1, _ := results[0].Node.(*model.Team)
		team2, _ := results[1].Node.(*model.Team)
		team3, _ := results[2].Node.(*model.Team)

		assert.Equal(t, "foo-team", team1.Slug)
		assert.Equal(t, "team-foo", team2.Slug)
		assert.Equal(t, "team-foo-bar", team3.Slug)
	})
}

func TestClient_GetTeam(t *testing.T) {
	ctx := context.Background()
	testLogger, _ := test.NewNullLogger()
	log := testLogger.WithContext(ctx)

	t.Run("team not found", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": {"teams": [{"slug": "team-1"}, {"slug": "team-2"}]}}`))
			},
		})
		team, err := teams.
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, false, errorsMeter(t), log).
			GetTeam(ctx, "foobar")

		assert.Nil(t, team)
		assert.EqualError(t, err, "team not found: foobar")
	})

	t.Run("team found", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": {"teams": [{"slug": "team-1"}, {"slug": "team-2"}]}}`))
			},
		})
		team, err := teams.
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, false, errorsMeter(t), log).
			GetTeam(ctx, "team-2")

		assert.Equal(t, "team-2", team.Slug)
		assert.NoError(t, err, err)
	})
}

func TestClient_GetGithubRepositories(t *testing.T) {
	ctx := context.Background()
	testLogger, _ := test.NewNullLogger()
	log := testLogger.WithContext(ctx)

	t.Run("team not found", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"errors": [{"message": "team not found"}],"data": null}`))
			},
		})
		repos, err := teams.
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, false, errorsMeter(t), log).
			GetGithubRepositories(ctx, "foobar")

		assert.Nil(t, repos)
		assert.EqualError(t, err, "team not found: foobar")
	})

	t.Run("team with no repos found", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data":{"team":{"gitHubRepositories":[]}}}`))
			},
		})
		repos, err := teams.
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, false, errorsMeter(t), log).
			GetGithubRepositories(ctx, "foobar")

		assert.NoError(t, err)
		assert.Len(t, repos, 0)
	})

	t.Run("team with repos found", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data":{"team":{"gitHubRepositories":[{"name":"org/repo-1"},{"name":"org/repo-2"}]}}}`))
			},
		})
		repos, err := teams.
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, false, errorsMeter(t), log).
			GetGithubRepositories(ctx, "foobar")

		assert.NoError(t, err)
		assert.Equal(t, "org/repo-1", repos[0].Name)
		assert.Equal(t, "org/repo-2", repos[1].Name)
	})
}

func TestClient_GetTeamMembers(t *testing.T) {
	ctx := context.Background()
	testLogger, _ := test.NewNullLogger()
	log := testLogger.WithContext(ctx)

	t.Run("team not found", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"errors": [{"message": "team not found"}],"data": null}`))
			},
		})
		members, err := teams.
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, false, errorsMeter(t), log).
			GetTeamMembers(ctx, "foobar")

		assert.Nil(t, members)
		assert.EqualError(t, err, "team not found: foobar")
	})

	t.Run("team found, no members", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data":{"team":{"members":[]}}}`))
			},
		})
		members, err := teams.
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, false, errorsMeter(t), log).
			GetTeamMembers(ctx, "foobar")

		assert.NoError(t, err)
		assert.Len(t, members, 0)
	})

	t.Run("team with members found", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data":{"team":{"members":[{"user":{"name":"Some User"}},{"user":{"name":"Some Other User"}}]}}}`))
			},
		})
		members, err := teams.
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, false, errorsMeter(t), log).
			GetTeamMembers(ctx, "foobar")

		assert.NoError(t, err)
		assert.Len(t, members, 2)
		assert.Equal(t, "Some User", members[0].User.Name)
		assert.Equal(t, "Some Other User", members[1].User.Name)
	})
}

func TestClient_Auth(t *testing.T) {
	ctx := context.Background()
	testLogger, _ := test.NewNullLogger()
	log := testLogger.WithContext(ctx)

	t.Run("IAP headers included", func(t *testing.T) {
		expectedAssertion := "assertion"
		expectedEmail := "email"
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				ass := r.Header.Get("X-Goog-IAP-JWT-Assertion")
				email := r.Header.Get("X-Goog-Authenticated-User-Email")
				assert.Equal(t, expectedAssertion, ass)
				assert.Equal(t, expectedEmail, email)
				w.Write([]byte(`{}`))
			},
		})
		ctx := auth.SetIAP(ctx, expectedAssertion, expectedEmail)
		_, err := teams.
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, true, errorsMeter(t), log).
			GetTeams(ctx)

		assert.Nil(t, err)
	})

	t.Run("Bearertoken included", func(t *testing.T) {
		teamsBackend := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				authHeader := r.Header.Get("Authorization")
				assert.Equal(t, "Bearer "+apiToken, authHeader)
				w.Write([]byte(`{}`))
			},
		})
		_, err := teams.
			New(config.Teams{Token: apiToken, Endpoint: teamsBackend.URL}, false, errorsMeter(t), log).
			GetTeams(ctx)

		assert.Nil(t, err)
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
