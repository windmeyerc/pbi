package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/bigquery"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type QueryRequest struct {
	Queries []Query `json:"queries"`
}

type Query struct {
	Query string `json:"query"`
}

type QueryResponse struct {
	Results []Result `json:"results"`
}

type Result struct {
	Tables []Table `json:"tables"`
}

type Table struct {
	Rows []map[string]interface{} `json:"rows"`
}

func getAccessToken(tenantID, clientID, clientSecret string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("scope", "https://analysis.windows.net/powerbi/api/.default")

	req, err := http.NewRequest("POST", fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID), strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get token: %s", body)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func executeQuery(accessToken, workspaceID, datasetID, query string) (*QueryResponse, error) {
	url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/groups/%s/datasets/%s/executeQueries", workspaceID, datasetID)

	queryReq := QueryRequest{
		Queries: []Query{{Query: query}},
	}

	jsonData, err := json.Marshal(queryReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to execute query: %s", body)
	}

	var queryResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, err
	}

	return &queryResp, nil
}

func main() {
	tenantID := os.Getenv("TENANT_ID")
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	workspaceID := os.Getenv("WORKSPACE_ID")
	datasetID := os.Getenv("DATASET_ID")

	if tenantID == "" || clientID == "" || clientSecret == "" || workspaceID == "" || datasetID == "" {
		log.Fatal("All environment variables (TENANT_ID, CLIENT_ID, CLIENT_SECRET, WORKSPACE_ID, DATASET_ID) must be set")
	}

	// Load queries from file
	queryData, err := os.ReadFile("queries.json")
	if err != nil {
		log.Fatal("Failed to read queries.json:", err)
	}

	var queries map[string]string
	if err := json.Unmarshal(queryData, &queries); err != nil {
		log.Fatal("Failed to parse queries.json:", err)
	}

	fmt.Println("Available queries:")
	for key, query := range queries {
		fmt.Printf("%s: %s\n", key, query)
	}

	fmt.Print("Select a query by number: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	selected := strings.TrimSpace(scanner.Text())

	query, exists := queries[selected]
	if !exists {
		log.Fatal("Invalid query selection")
	}

	accessToken, err := getAccessToken(tenantID, clientID, clientSecret)
	if err != nil {
		log.Fatal("Failed to get access token:", err)
	}

	result, err := executeQuery(accessToken, workspaceID, datasetID, query)
	if err != nil {
		log.Fatal("Failed to execute query:", err)
	}

	// Print results
	for _, res := range result.Results {
		for _, table := range res.Tables {
			for _, row := range table.Rows {
				fmt.Println(row)
			}
		}
	}

	// Optional: Insert results into BigQuery
	projectID := os.Getenv("GCP_PROJECT_ID")
	bqDatasetID := os.Getenv("BIGQUERY_DATASET_ID")
	bqTableID := os.Getenv("BIGQUERY_TABLE_ID")

	if projectID != "" && bqDatasetID != "" && bqTableID != "" {
		ctx := context.Background()
		var client *bigquery.Client
		var err2 error
		client, err2 = bigquery.NewClient(ctx, projectID)
		if err2 != nil {
			log.Printf("Failed to create BigQuery client: %v", err2)
			return
		}
		defer client.Close()

		// Assume first result, first table
		if len(result.Results) > 0 && len(result.Results[0].Tables) > 0 {
			rows := result.Results[0].Tables[0].Rows

			// Insert rows
			inserter := client.Dataset(bqDatasetID).Table(bqTableID).Inserter()
			if insertErr := inserter.Put(ctx, rows); insertErr != nil {
				log.Printf("Failed to insert rows into BigQuery: %v", insertErr)
				return
			}
			fmt.Println("Rows inserted into BigQuery successfully")
		}
	}
}
