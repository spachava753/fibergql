package main

import (
	"log"
	"net/http"

	todo "github.com/spachava753/fibergql/example/config"
	"github.com/spachava753/fibergql/graphql/handler"
	"github.com/spachava753/fibergql/graphql/playground"
)

func main() {
	http.Handle("/", playground.Handler("Todo", "/query"))
	http.Handle("/query", handler.NewDefaultServer(
		todo.NewExecutableSchema(todo.New()),
	))
	log.Fatal(http.ListenAndServe(":8081", nil))
}
