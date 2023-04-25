generate-graphql:
	go run github.com/99designs/gqlgen generate

local:
	go run ./cmd/console-backend/main.go -bind-host 127.0.0.1 -port 6969 --console-token secret --kubeconfig ./kubeconfig
