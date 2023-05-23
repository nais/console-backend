generate-graphql:
	go run github.com/99designs/gqlgen generate

local:
	go run ./cmd/console-backend/main.go --bind-host 127.0.0.1 --port 4242 --kubeconfig ./kubeconfig --run-as-user devuser@console.no --field-selector "metadata.namespace!=kube-system,metadata.namespace!=kyverno,metadata.namespace!=nais-system,metadata.namespace!=kimfoo,metadata.namespace!=nais-verification,metadata.namespace!=johnny"

setup: 
	gcloud secrets versions access latest --secret=console-backend-kubeconfig --project aura-dev-d9f5 > kubeconfig

linux-binary:
	GOOS=linux GOARCH=amd64 go build -o bin/console-backend ./cmd/console-backend/main.go
