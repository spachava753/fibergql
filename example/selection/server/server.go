package main

import (
	"log"
	"net/http"

	"github.com/spachava753/fibergql/example/selection"
	"github.com/spachava753/fibergql/graphql/handler"
	"github.com/spachava753/fibergql/graphql/playground"
)

func main() {
	http.Handle("/", playground.Handler("Selection Demo", "/query"))
	http.Handle("/query", handler.NewDefaultServer(selection.NewExecutableSchema(selection.Config{Resolvers: &selection.Resolver{}})))
	log.Fatal(http.ListenAndServe(":8086", nil))
}
