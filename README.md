# Power BI Semantic Model Query App

This Go application connects to a Power BI semantic model and allows executing predefined DAX queries at runtime.

## Prerequisites

- Go 1.19 or later
- Azure AD app registration with Power BI API permissions
- Google Cloud Platform (GCP) project with BigQuery enabled (optional, for writing results)
- Environment variables set for authentication and dataset details

## Setup

1. Register an Azure AD application and grant it Power BI API permissions.
2. (Optional) Set up GCP BigQuery:
   - Create a GCP project and enable BigQuery API.
   - Create a dataset and table in BigQuery (the table should have a compatible schema with the query results).
   - Ensure authentication is set up (e.g., via service account key or default credentials).
3. Set the following environment variables:
   - `TENANT_ID`: Your Azure tenant ID
   - `CLIENT_ID`: Azure AD app client ID
   - `CLIENT_SECRET`: Azure AD app client secret
   - `WORKSPACE_ID`: Power BI workspace ID
   - `DATASET_ID`: Power BI dataset ID
   - (Optional) `GCP_PROJECT_ID`: GCP project ID
   - (Optional) `BIGQUERY_DATASET_ID`: BigQuery dataset ID
   - (Optional) `BIGQUERY_TABLE_ID`: BigQuery table ID

## Usage

1. Build the application:
   ```bash
   go build
   ```

2. Run the application:
   ```bash
   ./pbi-app
   ```

3. Select a query from the list by entering the corresponding number.

4. The query results will be displayed in the console and optionally inserted into BigQuery if the GCP environment variables are set.

## Queries

The application reads queries from `queries.json`. Each query is a key-value pair where the key is the selection number and the value is the DAX query.

Example `queries.json`:

```json
{
  "1": "EVALUATE TOPN(10, 'Sales'[Product], 'Sales'[Quantity])",
  "2": "EVALUATE SUM('Sales'[Revenue])",
  "3": "EVALUATE AVERAGE('Sales'[Price])"
}
```

To add more queries, edit `queries.json` and add new entries.

## Testing

Run the unit tests with:
```bash
go test
```

The tests cover basic error handling for authentication and query execution, as well as JSON query loading.