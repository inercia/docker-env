package config

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestReplaceVars(t *testing.T) {
	vars := map[string]string{
		"NUM_DATABASES": "1",
		"NUM_WORKERS":   "5",
		"NUM_QUEUES":    "9",
	}

	s := "workers: $[NUM_WORKERS]"
	require.False(t, hasVars(s))

	s = "databases: $(NUM_DATABASES) workers: $(NUM_WORKERS) $(UNKNOWN)"
	require.True(t, hasVars(s))
	replaced, err := replaceVars(s, vars)
	require.Error(t, err, "replacement error")
	require.Equal(t, replaced, "databases: 1 workers: 5 $(UNKNOWN)")
}

func TestReplace(t *testing.T) {
	scs := spew.ConfigState{
		Indent:   "\t",
		SortKeys: true,
	}

	vars := map[string]string{
		"NUM_DATABASES": "1",
		"NUM_WORKERS":   "5",
		"NUM_QUEUES":    "9",
	}

	type testStruct struct {
		A string
		B string
		C int
	}

	a := testStruct{
		A: "databases: $(NUM_DATABASES)",
		B: "workers: $(NUM_WORKERS)",
		C: 10,
	}
	expected := testStruct{
		A: "databases: 1",
		B: "workers: 5",
		C: 10,
	}

	translated := ReplaceAllStringFunc(a, func(in string) string {
		replaced, err := replaceVars(in, vars)
		if err != nil {
			return in
		}
		return replaced
	})
	require.True(t, reflect.DeepEqual(translated, expected), scs.Sdump(translated))

	translated = ReplaceAllStringMap(a, vars)
	require.True(t, reflect.DeepEqual(translated, expected), scs.Sdump(translated))
}
