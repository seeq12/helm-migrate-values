package pkg

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// Every migration in a chain must see values with the same types Helm hands the first one:
// string-keyed maps (so sprig's set, merge and omit work on nested maps) and float64 numbers.
var chainedMigrationTestCases = []struct {
	name           string
	migrationsDir  string
	currentVersion int
	storedValues   string
	expected       map[string]interface{}
}{
	{
		// to-v5 calls set on .server.cache after to-v4 has run.
		name:           "set on nested map after an earlier migration",
		migrationsDir:  "test-migrations/nested-maps",
		currentVersion: 3,
		storedValues:   `{"mode": "standalone", "server": {"cpu": "500m", "logLevel": "debug", "cache": {"size": "2g", "ttl": "5m"}}, "network": {"poolSize": 8}}`,
		expected: map[string]interface{}{
			"mode": "standalone",
			"network": map[string]interface{}{
				"poolSize": float64(4),
			},
			"server": map[string]interface{}{
				"cpu":      "500m",
				"logLevel": "debug",
				"extraEnv": []interface{}{},
				"cache": map[string]interface{}{
					"size":    "2g",
					"ttl":     "5m",
					"enabled": true,
					"shards":  float64(2),
				},
			},
		},
	},
	{
		// to-v4 calls set on .network, which to-v3 has already re-decoded.
		name:           "set on nested map in the second migration of the chain",
		migrationsDir:  "test-migrations/nested-maps",
		currentVersion: 2,
		storedValues:   `{"mode": "legacy", "network": {"poolSize": 8, "timeout": 30}}`,
		expected: map[string]interface{}{
			"mode": "standalone",
			"network": map[string]interface{}{
				"poolSize": float64(4),
				"timeout":  float64(30),
			},
			"server": map[string]interface{}{
				"extraEnv": []interface{}{},
				"cache": map[string]interface{}{
					"enabled": true,
					"shards":  float64(2),
				},
			},
		},
	},
	{
		name:           "missing nested maps are created by a later migration",
		migrationsDir:  "test-migrations/nested-maps",
		currentVersion: 3,
		storedValues:   `{"mode": "standalone", "image": {"tag": "1.2.3"}}`,
		expected: map[string]interface{}{
			"mode": "standalone",
			"image": map[string]interface{}{
				"tag": "1.2.3",
			},
			"network": map[string]interface{}{
				"poolSize": float64(4),
			},
			"server": map[string]interface{}{
				"extraEnv": []interface{}{},
				"cache": map[string]interface{}{
					"enabled": true,
					"shards":  float64(2),
				},
			},
		},
	},
	{
		name:           "null nested map is replaced by a later migration",
		migrationsDir:  "test-migrations/nested-maps",
		currentVersion: 3,
		storedValues:   `{"mode": "standalone", "server": {"cpu": "500m", "cache": null}}`,
		expected: map[string]interface{}{
			"mode": "standalone",
			"network": map[string]interface{}{
				"poolSize": float64(4),
			},
			"server": map[string]interface{}{
				"cpu":      "500m",
				"extraEnv": []interface{}{},
				"cache": map[string]interface{}{
					"enabled": true,
					"shards":  float64(2),
				},
			},
		},
	},
	{
		name:           "single migration returns string-keyed maps",
		migrationsDir:  "test-migrations/nested-maps",
		currentVersion: 4,
		storedValues:   `{"mode": "standalone", "server": {"cache": {"size": "2g"}}, "network": {"poolSize": 4}}`,
		expected: map[string]interface{}{
			"mode": "standalone",
			"network": map[string]interface{}{
				"poolSize": float64(4),
			},
			"server": map[string]interface{}{
				"extraEnv": []interface{}{},
				"cache": map[string]interface{}{
					"size":    "2g",
					"enabled": true,
					"shards":  float64(2),
				},
			},
		},
	},
	{
		name:           "set on maps inside a list after an earlier migration",
		migrationsDir:  "test-migrations/list-of-maps",
		currentVersion: 1,
		storedValues:   `{"sidecars": [{"name": "a"}, {"name": "b"}]}`,
		expected: map[string]interface{}{
			"sidecars": []interface{}{
				map[string]interface{}{"name": "a", "enabled": true},
				map[string]interface{}{"name": "b", "enabled": true},
			},
		},
	},
	{
		name:           "numbers are float64 in the first migration",
		migrationsDir:  "test-migrations/number-types",
		currentVersion: 2,
		storedValues:   `{"replicas": 3}`,
		expected: map[string]interface{}{
			"replicas":     float64(3),
			"replicasKind": "float64",
		},
	},
	{
		name:           "numbers are float64 after an earlier migration",
		migrationsDir:  "test-migrations/number-types",
		currentVersion: 1,
		storedValues:   `{"replicas": 3}`,
		expected: map[string]interface{}{
			"replicas":     float64(3),
			"replicasKind": "float64",
		},
	},
}

func TestMigrator_ChainedMigrations(t *testing.T) {
	for _, tc := range chainedMigrationTestCases {
		t.Run(tc.name, func(t *testing.T) {
			is := assert.New(t)
			req := require.New(t)

			migrated, err := MigrateFromPath(helmStoredValues(t, tc.storedValues), tc.currentVersion, nil, tc.migrationsDir, *NewLogger(false))
			req.NoError(err)

			is.Equal(tc.expected, migrated)
		})
	}
}

func TestMigrator_DoesNotModifyCurrentConfig(t *testing.T) {
	storedValues := `{"server": {"cache": {"size": "2g"}}}`
	currentConfig := helmStoredValues(t, storedValues)

	_, err := MigrateFromPath(currentConfig, 4, nil, "test-migrations/nested-maps", *NewLogger(false))
	require.NoError(t, err)

	assert.Equal(t, helmStoredValues(t, storedValues), currentConfig)
}

// helmStoredValues decodes values the way Helm's storage drivers return a release's Config: through
// encoding/json, so nested maps are map[string]interface{} and numbers are float64.
func helmStoredValues(t *testing.T, storedValues string) map[string]interface{} {
	t.Helper()

	var values map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(storedValues), &values))
	return values
}
