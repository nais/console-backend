generate-graphql:
	go run github.com/99designs/gqlgen generate

local-nav:
	go run ./cmd/console-backend/main.go --bind-host 127.0.0.1 --port 4242 --kubernetes-clusters "dev-gcp,prod-gcp" --run-as-user johnny.horvi@nav.no --teams-endpoint="http://localhost:8181/query" --hookd-endpoint="http://localhost:8282" --tenant="nav"

local:
	go run ./cmd/console-backend/main.go --bind-host 127.0.0.1 --port 4242 --kubernetes-clusters "ci,dev" --run-as-user devuser@console.no --teams-endpoint="http://teams.local.nais.io/query" --hookd-endpoint="http://hookd.local.nais.io" --field-selector "metadata.namespace!=kube-system,metadata.namespace!=kyverno,metadata.namespace!=nais-system,metadata.namespace!=kimfoo,metadata.namespace!=nais-verification,metadata.namespace!=johnny"

test:
	go run ./cmd/console-backend/main.go --bind-host 127.0.0.1 --port 4242 --kubernetes-projects "nais-ci-2a63,nais-dev-cdea" --teams-endpoint="http://teams.local.nais.io/query" --hookd-endpoint="http://hookd.local.nais.io" --log-level debug

setup: 
	gcloud secrets versions access latest --secret=console-backend-kubeconfig --project aura-dev-d9f5 > kubeconfig

linux-binary:
	GOOS=linux GOARCH=amd64 go build -o bin/console-backend ./cmd/console-backend/main.go
