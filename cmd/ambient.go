package cmd

import (
	// Import and ignore the ambient imports listed below so dependency managers
	// don't prune unused code for us. Both lists should be kept in sync.
	_ "github.com/spachava753/fibergql/graphql"
	_ "github.com/spachava753/fibergql/graphql/introspection"
	_ "github.com/vektah/gqlparser/v2"
	_ "github.com/vektah/gqlparser/v2/ast"
)
