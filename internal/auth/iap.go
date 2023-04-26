package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/idtoken"
)

type contextKey int

const contextEmail contextKey = 1

func StaticUser(user string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), contextEmail, user)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}

func ValidateIAPJWT(aud string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			iapJWT := r.Header.Get("X-Goog-IAP-JWT-Assertion")

			payload, err := idtoken.Validate(r.Context(), iapJWT, aud)
			if err != nil {
				http.Error(w, "Invalid JWT token", http.StatusUnauthorized)
				return
			}

			if time.Unix(payload.IssuedAt, 0).After(time.Now().Add(30 * time.Second)) {
				http.Error(w, "JWT token is in the future", http.StatusUnauthorized)
				return
			}

			if payload.Issuer != "https://cloud.google.com/iap" {
				http.Error(w, "Invalid JWT token issuer", http.StatusUnauthorized)
				return
			}

			email := r.Header.Get("X-Goog-Authenticated-User-Email")
			_, email, _ = strings.Cut(email, ":")

			ctx := context.WithValue(r.Context(), contextEmail, email)
			r = r.WithContext(ctx)

			h.ServeHTTP(w, r)
		})
	}
}

func GetEmail(ctx context.Context) (string, error) {
	email, ok := ctx.Value(contextEmail).(string)
	if !ok || email == "" {
		return "", fmt.Errorf("no email in context")
	}
	return email, nil
}
