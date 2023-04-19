package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/graph"
)

const (
	defaultPort = "8080"
	audience    = ""
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Add the IAP validation middleware.
	// If the IAP audience is not set, we stop the server with a fatal error
	// unless the insecure-skip-proxy flag is set.
	iapMW := auth.ValidateIAPJWT(audience)
	if audience == "" {
		// if !cfg.InsecureSkipProxy {
		// 	log.Fatal("IAP audience must be set")
		// }
		iapMW = auth.InsecureValidateMW
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", iapMW(srv))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, nil))
}
