generate-graphql:
	go run github.com/99designs/gqlgen generate

local:
	go run ./cmd/console-backend/main.go
