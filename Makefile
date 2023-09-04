generate-graphql:
	go run github.com/99designs/gqlgen generate

setup: 
	gcloud secrets versions access latest --secret=console-backend-kubeconfig --project aura-dev-d9f5 > kubeconfig

linux-binary:
	GOOS=linux GOARCH=amd64 go build -o bin/console-backend ./cmd/console-backend/main.go

portforward-hookd:
	kubectl port-forward -n nais-system --context nav-management svc/hookd 8282:80

portforward-teams:
	kubectl port-forward -n nais-system --context nav-management svc/teams-backend 8181:80

local-nav:
	TEAMS_TOKEN="$(shell kubectl get secret console-backend --context nav-management -n nais-system -ojsonpath='{.data.TEAMS_TOKEN}' | base64 -D)" \
	HOOKD_PSK="$(shell kubectl get secret console-backend --context nav-management -n nais-system -ojsonpath='{.data.HOOKD_PSK}' | base64 -D)" \
	KUBERNETES_CLUSTERS_STATIC="dev-fss|apiserver.dev-fss.nais.io|$(shell kubectl get secret --context dev-fss --namespace nais-system console-backend -ojsonpath='{ .data.token }' | base64 -D)" \
	go run ./cmd/console-backend/main.go --bind-host 127.0.0.1 --port 4242 --kubernetes-clusters "dev-gcp,prod-gcp" --run-as-user johnny.horvi@nav.no --teams-endpoint="http://localhost:8181/query" --hookd-endpoint="http://localhost:8282" --tenant="nav" --field-selector "metadata.namespace!=kube-system,metadata.namespace!=kyverno,metadata.namespace!=nais-system,metadata.namespace!=kimfoo,metadata.namespace!=johnny"

local:
	go run ./cmd/console-backend/main.go --bind-host 127.0.0.1 --port 4242 --kubernetes-clusters "ci,dev" --run-as-user devuser@console.no --teams-endpoint="http://teams.local.nais.io/query" --hookd-endpoint="http://hookd.local.nais.io" --field-selector "metadata.namespace!=kube-system,metadata.namespace!=kyverno,metadata.namespace!=nais-system,metadata.namespace!=kimfoo,metadata.namespace!=johnny"

test:
	go test ./... 

check:
	go run golang.org/x/vuln/cmd/govulncheck ./...