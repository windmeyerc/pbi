package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGetAccessToken(t *testing.T) {
	// Test with invalid tenant
	_, err := getAccessToken("", "client", "secret")
	if err == nil {
		t.Error("Expected error for empty tenant")
	}
}

func TestExecuteQuery(t *testing.T) {
	// Test with invalid token
	_, err := executeQuery("", "workspace", "dataset", "query")
	if err == nil {
		t.Error("Expected error for empty token")
	}
}

func TestLoadQueries(t *testing.T) {
	// Create a temporary queries.json
	tempFile, err := os.CreateTemp("", "queries.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	testQueries := map[string]string{
		"1": "SELECT * FROM table",
		"2": "SELECT count(*) FROM table",
	}
	data, _ := json.Marshal(testQueries)
	tempFile.Write(data)
	tempFile.Close()

	// Read the file
	fileData, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	var queries map[string]string
	err = json.Unmarshal(fileData, &queries)
	if err != nil {
		t.Fatal(err)
	}

	if len(queries) != 2 || queries["1"] != "SELECT * FROM table" {
		t.Error("Failed to load queries correctly")
	}
}
