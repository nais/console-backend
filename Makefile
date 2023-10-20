.PHONY: all

all: generate fmt test check console-backend

generate: generate-sql generate-graphql generate-mocks

generate-sql:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc generate -f .sqlc.yaml
	go run github.com/sqlc-dev/sqlc/cmd/sqlc vet -f .sqlc.yaml
	go run mvdan.cc/gofumpt -w ./internal/database/gensql


generate-graphql:
	go run github.com/99designs/gqlgen generate
	go run mvdan.cc/gofumpt -w ./internal/graph

generate-mocks:
	go run github.com/vektra/mockery/v2
	find internal -type f -name "mock_*.go" -exec go run mvdan.cc/gofumpt -w {} \;

setup: 
	gcloud secrets versions access latest --secret=console-backend-kubeconfig --project aura-dev-d9f5 > kubeconfig

console-backend:
	go build -o bin/console-backend ./cmd/console-backend/main.go

portforward-hookd:
	kubectl port-forward -n nais-system --context nav-management-v2 svc/hookd 8282:80

portforward-teams:
	kubectl port-forward -n nais-system --context nav-management-v2 svc/teams-backend 8181:80

local-nav:
	HOOKD_ENDPOINT="http://localhost:8282" \
	HOOKD_PSK="$(shell kubectl get secret console-backend --context nav-management-v2 -n nais-system -ojsonpath='{.data.HOOKD_PSK}' | base64 --decode)" \
	KUBERNETES_CLUSTERS="dev-gcp,prod-gcp" \
	KUBERNETES_CLUSTERS_STATIC="dev-fss|apiserver.dev-fss.nais.io|$(shell kubectl get secret --context dev-fss --namespace nais-system console-backend -ojsonpath='{ .data.token }' | base64 --decode)" \
	KUBERNETES_FIELD_SELECTOR="metadata.namespace!=kube-system,metadata.namespace!=kyverno,metadata.namespace!=nais-system,metadata.namespace!=kimfoo,metadata.namespace!=johnny,metadata.namespace!=nais" \
	LISTEN_ADDRESS=":4242" \
	LOG_FORMAT="text" \
	LOG_LEVEL="debug" \
	RUN_AS_USER="johnny.horvi@nav.no" \
	TEAMS_ENDPOINT="http://localhost:8181/query" \
	TEAMS_TOKEN="$(shell kubectl get secret console-backend --context nav-management-v2 -n nais-system -ojsonpath='{.data.TEAMS_TOKEN}' | base64 --decode)" \
	TENANT="nav" \
	go run ./cmd/console-backend/main.go

local:
	HOOKD_ENDPOINT="http://hookd.local.nais.io" \
	KUBERNETES_CLUSTERS="ci,dev" \
	KUBERNETES_FIELD_SELECTOR="metadata.namespace!=kube-system,metadata.namespace!=kyverno,metadata.namespace!=nais-system,metadata.namespace!=kimfoo,metadata.namespace!=johnny" \
	LISTEN_ADDRESS=":4242" \
	LOG_FORMAT="text" \
	LOG_LEVEL="debug" \
	RUN_AS_USER="devuser@console.no" \
	TEAMS_ENDPOINT="http://teams.local.nais.io/query" \
	go run ./cmd/console-backend/main.go

test:
	go test ./... -v

check:
	go run honnef.co/go/tools/cmd/staticcheck ./...
	go run golang.org/x/vuln/cmd/govulncheck ./...

fmt:
	go run mvdan.cc/gofumpt -w ./
