package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Initialize an in-memory SQLite database for testing
func initTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// Create the resources table schema
	_, err = db.Exec(`
		CREATE TABLE resources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid VARCHAR(100) NOT NULL UNIQUE,
			name VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	// Insert sample data into the table using fixed times
	sampleData := []struct {
		UUID, Name string
		CreatedAt  time.Time
	}{
		{"uuid-1", "Resource 1", parseTime("2024-09-25T10:00:00Z")},
		{"uuid-2", "Resource 2", parseTime("2024-09-25T10:05:00Z")},
		{"uuid-3", "Resource 3", parseTime("2024-09-25T10:10:00Z")},
		{"uuid-4", "Resource 4", parseTime("2024-09-25T10:15:00Z")},
		{"uuid-5", "Resource 5", parseTime("2024-09-25T10:20:00Z")},
	}

	for _, data := range sampleData {
		_, err := db.Exec(`INSERT INTO resources (uuid, name, created_at) VALUES (?, ?, ?)`, data.UUID, data.Name, data.CreatedAt)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

// parseTime is a helper function to parse a time string into a time.Time object.
func parseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		log.Fatalf("Failed to parse time: %v", err)
	}
	return t
}

// FetchResourceByName fetches the resource by its name for verification.
func fetchResourceByName(db *sql.DB, name string) ([]*Resource, error) {
	query := `SELECT id, uuid, name, created_at, updated_at FROM resources WHERE name = ? ORDER BY created_at`
	rows, err := db.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*Resource
	for rows.Next() {
		var r Resource
		if err := rows.Scan(&r.ID, &r.UUID, &r.Name, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		resources = append(resources, &r)
	}

	return resources, nil
}

// TestResourceKeysetPage tests the ResourceKeysetPage function.
func TestResourceKeysetPage(t *testing.T) {
	db, err := initTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Fetch expected results for validation
	resource2, err := fetchResourceByName(db, "Resource 2")
	if err != nil {
		t.Fatalf("Failed to fetch Resource 2: %v", err)
	}

	resource3, err := fetchResourceByName(db, "Resource 3")
	if err != nil {
		t.Fatalf("Failed to fetch Resource 3: %v", err)
	}

	resource4, err := fetchResourceByName(db, "Resource 4")
	if err != nil {
		t.Fatalf("Failed to fetch Resource 4: %v", err)
	}

	tests := []struct {
		name        string
		column      string
		key         interface{}
		limit       int
		order       string
		filters     map[string]interface{}
		expected    []*Resource
		expectError bool
	}{
		{
			name:   "Valid ASC order with no filters",
			column: "created_at",
			key:    parseTime("2024-09-25T09:55:00Z"),
			limit:  3,
			order:  "ASC",
			filters: map[string]interface{}{
				"name": "Resource 2",
			},
			expected:    resource2,
			expectError: false,
		},
		{
			name:   "Valid DESC order with no filters",
			column: "created_at",
			key:    parseTime("2024-09-25T10:30:00Z"),
			limit:  2,
			order:  "DESC",
			filters: map[string]interface{}{
				"name": "Resource 3",
			},
			expected:    resource3,
			expectError: false,
		},
		{
			name:   "Valid order with IN clause filter",
			column: "created_at",
			key:    parseTime("2024-09-25T10:07:00Z"),
			limit:  3,
			order:  "ASC",
			filters: map[string]interface{}{
				"name": []string{"Resource 3", "Resource 4"},
			},
			expected:    append(resource3, resource4...),
			expectError: false,
		},
		{
			name:   "Invalid order value",
			column: "created_at",
			key:    parseTime("2024-09-25T09:55:00Z"),
			limit:  2,
			order:  "INVALID",
			filters: map[string]interface{}{
				"name": "Resource 1",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:   "Empty column name",
			column: "",
			key:    parseTime("2024-09-25T09:55:00Z"),
			limit:  2,
			order:  "ASC",
			filters: map[string]interface{}{
				"name": "Resource 1",
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, _, err := ResourceKeysetPage(ctx, db, tt.column, tt.key, tt.limit, tt.order, tt.filters)
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
				return
			}

			if !equalResourceSlices(result, tt.expected) {
				t.Errorf("Expected result: %+v, got: %+v", tt.expected, result)
			}
		})
	}
}
func TestResourceKeysetPagePagination(t *testing.T) {
	db, err := initTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// First Page: Get the first 2 resources ordered by created_at ASC
	firstPage, pageKey, err := ResourceKeysetPage(ctx, db, "created_at", parseTime("2024-09-25T09:55:00Z"), 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get first page: %v", err)
	}

	expectedFirstPage := []*Resource{
		{ID: 1, UUID: "uuid-1", Name: "Resource 1", CreatedAt: parseTime("2024-09-25T10:00:00Z")},
		{ID: 2, UUID: "uuid-2", Name: "Resource 2", CreatedAt: parseTime("2024-09-25T10:05:00Z")},
	}

	if !equalResourceSlices(firstPage, expectedFirstPage) {
		t.Errorf("Expected first page: %+v, got: %+v", printResources(expectedFirstPage), printResources(firstPage))
	}

	lastCreatedAt := pageKey.CreatedAt

	// Second Page: Get the next 2 resources ordered by created_at ASC
	secondPage, pageKey, err := ResourceKeysetPage(ctx, db, "created_at", lastCreatedAt, 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get second page: %v", err)
	}

	expectedSecondPage := []*Resource{
		{ID: 3, UUID: "uuid-3", Name: "Resource 3", CreatedAt: parseTime("2024-09-25T10:10:00Z")},
		{ID: 4, UUID: "uuid-4", Name: "Resource 4", CreatedAt: parseTime("2024-09-25T10:15:00Z")},
	}

	if !equalResourceSlices(secondPage, expectedSecondPage) {
		t.Errorf("Expected second page: %+v, got: %+v", printResources(expectedSecondPage), printResources(secondPage))
	}

	// Use the last object's CreatedAt value as the key for the next page
	lastCreatedAt = pageKey.CreatedAt

	// Third Page: Get the remaining resources
	thirdPage, _, err := ResourceKeysetPage(ctx, db, "created_at", lastCreatedAt, 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get third page: %v", err)
	}

	expectedThirdPage := []*Resource{
		{ID: 5, UUID: "uuid-5", Name: "Resource 5", CreatedAt: parseTime("2024-09-25T10:20:00Z")},
	}

	if !equalResourceSlices(thirdPage, expectedThirdPage) {
		t.Errorf("Expected third page: %+v, got: %+v", printResources(expectedThirdPage), printResources(thirdPage))
	}

	// ---------------------------------
	// New Test: Pagination Based on Name
	// ---------------------------------

	// First Page: Get the first 2 resources ordered by name ASC
	firstPageByName, pageKey, err := ResourceKeysetPage(ctx, db, "name", "Resource 0", 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get first page by name: %v", err)
	}

	expectedFirstPageByName := []*Resource{
		{ID: 1, UUID: "uuid-1", Name: "Resource 1", CreatedAt: parseTime("2024-09-25T10:00:00Z")},
		{ID: 2, UUID: "uuid-2", Name: "Resource 2", CreatedAt: parseTime("2024-09-25T10:05:00Z")},
	}

	if !equalResourceSlices(firstPageByName, expectedFirstPageByName) {
		t.Errorf("Expected first page by name: %+v, got: %+v", printResources(expectedFirstPageByName), printResources(firstPageByName))
	}

	// Use the last object's Name value as the key for the next page
	lastName := pageKey.Name

	// Second Page: Get the next 2 resources ordered by name ASC
	secondPageByName, pageKey, err := ResourceKeysetPage(ctx, db, "name", lastName, 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get second page by name: %v", err)
	}

	expectedSecondPageByName := []*Resource{
		{ID: 3, UUID: "uuid-3", Name: "Resource 3", CreatedAt: parseTime("2024-09-25T10:10:00Z")},
		{ID: 4, UUID: "uuid-4", Name: "Resource 4", CreatedAt: parseTime("2024-09-25T10:15:00Z")},
	}

	if !equalResourceSlices(secondPageByName, expectedSecondPageByName) {
		t.Errorf("Expected second page by name: %+v, got: %+v", printResources(expectedSecondPageByName), printResources(secondPageByName))
	}

	// Use the last object's Name value as the key for the next page
	lastName = pageKey.Name

	// Third Page: Get the remaining resources ordered by name ASC
	thirdPageByName, _, err := ResourceKeysetPage(ctx, db, "name", lastName, 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get third page by name: %v", err)
	}

	expectedThirdPageByName := []*Resource{
		{ID: 5, UUID: "uuid-5", Name: "Resource 5", CreatedAt: parseTime("2024-09-25T10:20:00Z")},
	}

	if !equalResourceSlices(thirdPageByName, expectedThirdPageByName) {
		t.Errorf("Expected third page by name: %+v, got: %+v", printResources(expectedThirdPageByName), printResources(thirdPageByName))
	}
}

// Utility function to print the Resource slice for debugging.
func printResources(resources []*Resource) string {
	var output string
	for _, r := range resources {
		output += fmt.Sprintf("{ID: %d, UUID: %s, Name: %s, CreatedAt: %s}\n", r.ID, r.UUID, r.Name, r.CreatedAt)
	}
	return output
}

// Utility function to compare two slices of *Resource (used in reflect.DeepEqual).
func equalResourceSlices(a, b []*Resource) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !equalResources(a[i], b[i]) {
			return false
		}
	}
	return true
}

// Utility function to compare two *Resource instances (used in equalResourceSlices).
func equalResources(a, b *Resource) bool {
	return a.ID == b.ID && a.UUID == b.UUID && a.Name == b.Name && a.CreatedAt.Equal(b.CreatedAt)
}
