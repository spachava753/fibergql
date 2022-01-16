package api

import (
	"testing"

	"github.com/spachava753/fibergql/codegen/config"
	"github.com/spachava753/fibergql/plugin"
	"github.com/spachava753/fibergql/plugin/federation"
	"github.com/spachava753/fibergql/plugin/modelgen"
	"github.com/spachava753/fibergql/plugin/resolvergen"
	"github.com/stretchr/testify/require"
)

type testPlugin struct{}

// Name returns the plugin name
func (t *testPlugin) Name() string {
	return "modelgen"
}

// MutateConfig mutates the configuration
func (t *testPlugin) MutateConfig(_ *config.Config) error {
	return nil
}

func TestReplacePlugin(t *testing.T) {
	t.Run("replace plugin if exists", func(t *testing.T) {
		pg := []plugin.Plugin{
			federation.New(),
			modelgen.New(),
			resolvergen.New(),
		}

		expectedPlugin := &testPlugin{}
		ReplacePlugin(expectedPlugin)(config.DefaultConfig(), &pg)

		require.EqualValues(t, federation.New(), pg[0])
		require.EqualValues(t, expectedPlugin, pg[1])
		require.EqualValues(t, resolvergen.New(), pg[2])
	})

	t.Run("add plugin if doesn't exist", func(t *testing.T) {
		pg := []plugin.Plugin{
			federation.New(),
			resolvergen.New(),
		}

		expectedPlugin := &testPlugin{}
		ReplacePlugin(expectedPlugin)(config.DefaultConfig(), &pg)

		require.EqualValues(t, federation.New(), pg[0])
		require.EqualValues(t, resolvergen.New(), pg[1])
		require.EqualValues(t, expectedPlugin, pg[2])
	})
}
