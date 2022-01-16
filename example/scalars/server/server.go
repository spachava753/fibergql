package main

import (
	"log"
	"net/http"

	"github.com/spachava753/fibergql/example/scalars"
	"github.com/spachava753/fibergql/graphql/handler"
	"github.com/spachava753/fibergql/graphql/playground"
)

func main() {
	http.Handle("/", playground.Handler("Starwars", "/query"))
	http.Handle("/query", handler.NewDefaultServer(scalars.NewExecutableSchema(scalars.Config{Resolvers: &scalars.Resolver{}})))

	log.Fatal(http.ListenAndServe(":8084", nil))
}
