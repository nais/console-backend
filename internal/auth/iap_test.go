package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nais/console-backend/internal/auth"
	"github.com/stretchr/testify/assert"
)

const (
	user = "user@example.com"
	aud  = "some audience"
)

func TestStaticUser(t *testing.T) {
	ctx := context.Background()

	t.Run("static user", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			emailFromContext, err := auth.GetEmail(r.Context())
			assert.Equal(t, user, emailFromContext)
			assert.NoError(t, err)
		})
		auth.StaticUser(user, handler).ServeHTTP(httptest.NewRecorder(), getRequest(t, ctx))
	})
}

func TestValidateIAPJWT(t *testing.T) {
	ctx := context.Background()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "should not be executed")
	})

	t.Run("invalid JWT token", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := getRequest(t, ctx)
		req.Header.Set("X-Goog-IAP-JWT-Assertion", "invalid JWT token")
		auth.ValidateIAPJWT(aud, handler).ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "Invalid JWT token")
	})

	t.Run("aud mismatch", func(t *testing.T) {
		// generated on jwt.io
		jwtWithIncorrectAudience := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJmb29iYXIiLCJpYXQiOjE1MTYyMzkwMjJ9.AAZSAganBukxb13XnhxWacEUE9Tv2RnZYIrcdW0pVCwHuDVkcCIkPXC53tIKgcNiLZp48OkjA0zU8mDSyV8gCh6_I8Emr2DCC4FIbGxPlL5wbKbgtstDADnitSL0H79Mp0Ko3IPb3tbbymq2ntN0N49jpnds_iMele9LWFPlggUSUWsJy4Wh2RU3EU8cxHc4jEuYztHzkr1u1lpUkHIgRx909Q04nCkzyZLymApVOyHYC-CGh7OhetwHMP1nJEub8KUZc321IBeX9C9ZsdFpfW3C1Y6yYUCmDKPoL4Kp_Ufphi1vgT7vBlUvoxWgihz8s75Cug_7-JLQ5oMJDClWrQ"

		recorder := httptest.NewRecorder()
		req := getRequest(t, ctx)
		req.Header.Set("X-Goog-IAP-JWT-Assertion", jwtWithIncorrectAudience)
		auth.ValidateIAPJWT(aud, handler).ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "Invalid JWT token")
	})
}

func getRequest(t *testing.T, ctx context.Context) *http.Request {
	t.Helper()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	return req
}
