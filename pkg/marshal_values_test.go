package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chartutil"
	"testing"
)

// Values reach the output as float64, so whole numbers must still be written as integers, and reading the output
// back must give the value Helm would have.
//
// These cases are why values are written with sigs.k8s.io/yaml, the encoder Helm uses, rather than gopkg.in/yaml.v2.
// yaml.v2 formats a float64 with strconv's shortest 'g' form, which switches to exponent notation at 1e+06, so a
// whole number of a million or more was written as 1e+06 or 1.2345678901e+10. Helm reads those back as the same
// number, but they are hard to read and diff, and YAML 1.1 parsers that require a '.' in floats, such as PyYAML,
// read 1e+06 as a string. sigs.k8s.io/yaml goes through encoding/json, which writes whole float64 values below
// 1e21 as plain integers. Cases marked "yaml.v2 wrote" are the ones the switch fixed; the rest are written the same
// by both and guard against the switch changing anything else.
var marshalValuesTestCases = []struct {
	name     string
	value    interface{}
	expected string
	readBack interface{}
}{
	{name: "zero", value: float64(0), expected: "0", readBack: float64(0)},
	{name: "small integer", value: float64(3), expected: "3", readBack: float64(3)},
	{name: "negative integer", value: float64(-42), expected: "-42", readBack: float64(-42)},
	// yaml.v2 wrote 1e+06
	{name: "million", value: float64(1000000), expected: "1000000", readBack: float64(1000000)},
	// yaml.v2 wrote 1.2345678901e+10
	{name: "integer above 32 bits", value: float64(12345678901), expected: "12345678901", readBack: float64(12345678901)},
	// yaml.v2 wrote 9.007199254740992e+15
	{name: "largest exact float64 integer", value: float64(1 << 53), expected: "9007199254740992", readBack: float64(1 << 53)},
	{name: "whole float", value: float64(2.0), expected: "2", readBack: float64(2)},
	{name: "fraction", value: float64(0.5), expected: "0.5", readBack: float64(0.5)},
	{name: "negative fraction", value: float64(-1.25), expected: "-1.25", readBack: float64(-1.25)},
	{name: "small fraction", value: float64(0.000015), expected: "1.5e-05", readBack: float64(0.000015)},
	{name: "int from a library caller", value: 1000000, expected: "1000000", readBack: float64(1000000)},
	{name: "numeric string stays a string", value: "1000000", expected: `"1000000"`, readBack: "1000000"},
	{name: "version string stays a string", value: "1.10", expected: `"1.10"`, readBack: "1.10"},
}

func TestMarshalValues_Numbers(t *testing.T) {
	for _, tc := range marshalValuesTestCases {
		t.Run(tc.name, func(t *testing.T) {
			is := assert.New(t)
			req := require.New(t)

			values := map[string]interface{}{"value": tc.value}
			expected := "value: " + tc.expected + "\n"

			out, err := MarshalValues(values)
			req.NoError(err)
			is.Equal(expected, string(out))

			is.Equal(expected, toYaml(values), "template toYaml")

			readBack, err := chartutil.ReadValues(out)
			req.NoError(err)
			is.Equal(tc.readBack, readBack["value"])
		})
	}
}

// Every migration's output is written with toYaml and read back before the next one runs, so a number that was
// written as 1e+06 between steps used to reach the final output in exponent notation too.
func TestMarshalValues_NumbersAfterChainedMigrations(t *testing.T) {
	migrated, err := MigrateFromPath(helmStoredValues(t, `{"replicas": 1000000}`), 1, nil, "test-migrations/number-types", *NewLogger(false))
	require.NoError(t, err)

	out, err := MarshalValues(migrated)
	require.NoError(t, err)

	assert.Equal(t, "replicas: 1000000\nreplicasKind: float64\n", string(out))
}
