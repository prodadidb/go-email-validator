package test

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// DefaultDepFixtureFile is a name of test file
const DefaultDepFixtureFile = "dep_fixture_test.json"

// DepPresentations returns structs from json test file
func DepPresentations(t *testing.T, result interface{}, fp string) {
	if fp == "" {
		fp = DefaultDepFixtureFile
	}

	fp, err := filepath.Abs(fp)
	require.Nil(t, err)
	jsonFile, err := os.Open(fp)
	require.Nil(t, err)
	defer func() {
		_ = jsonFile.Close()
	}()

	byteValue, err := io.ReadAll(jsonFile)
	require.Nil(t, err)

	err = json.Unmarshal(byteValue, &result)
	require.Nil(t, err)
}
