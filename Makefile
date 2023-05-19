generate-graphql:
	go run github.com/99designs/gqlgen generate

local:
	go run ./cmd/console-backend/main.go --bind-host 127.0.0.1 --port 4242 --kubeconfig ./kubeconfig --run-as-user frode@dev-nais.io

setup: 
	gcloud secrets versions access latest --secret=console-backend-kubeconfig --project aura-dev-d9f5 > kubeconfig
