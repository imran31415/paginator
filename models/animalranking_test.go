package models

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestAnimalRankingsPagination(t *testing.T) {
	// Initialize the in-memory database with the animal_rankings table and data
	db, err := initAnimalRankingsTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// First Page: Get the first 2 animal rankings ordered by rank ASC
	firstPage, pageKey, err := AnimalRankingKeysetPage(ctx, db, "rank", 0, 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get first page: %v", err)
	}

	expectedFirstPage := []*AnimalRanking{
		{ID: 1, Rank: 1, Name: "Lion", CreatedAt: parseTime("2024-09-25T10:00:00Z")},
		{ID: 2, Rank: 2, Name: "Tiger", CreatedAt: parseTime("2024-09-25T10:05:00Z")},
	}

	if !equalAnimalRankingSlices(firstPage, expectedFirstPage) {
		t.Errorf("Expected first page: %+v, got: %+v", printAnimalRankings(expectedFirstPage), printAnimalRankings(firstPage))
	}

	// Use the last object's rank value as the key for the next page
	lastRank := pageKey.Rank

	// Second Page: Get the next 2 animal rankings ordered by rank ASC
	secondPage, pageKey, err := AnimalRankingKeysetPage(ctx, db, "rank", lastRank, 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get second page: %v", err)
	}

	expectedSecondPage := []*AnimalRanking{
		{ID: 3, Rank: 3, Name: "Elephant", CreatedAt: parseTime("2024-09-25T10:10:00Z")},
		{ID: 4, Rank: 4, Name: "Leopard", CreatedAt: parseTime("2024-09-25T10:15:00Z")},
	}

	if !equalAnimalRankingSlices(secondPage, expectedSecondPage) {
		t.Errorf("Expected second page: %+v, got: %+v", printAnimalRankings(expectedSecondPage), printAnimalRankings(secondPage))
	}

	// Use the last object's rank value as the key for the next page
	lastRank = pageKey.Rank

	// Third Page: Get the remaining animal rankings ordered by rank ASC
	thirdPage, _, err := AnimalRankingKeysetPage(ctx, db, "rank", lastRank, 2, "ASC", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get third page: %v", err)
	}

	expectedThirdPage := []*AnimalRanking{
		{ID: 5, Rank: 5, Name: "Wolf", CreatedAt: parseTime("2024-09-25T10:20:00Z")},
	}

	if !equalAnimalRankingSlices(thirdPage, expectedThirdPage) {
		t.Errorf("Expected third page: %+v, got: %+v", printAnimalRankings(expectedThirdPage), printAnimalRankings(thirdPage))
	}
}

// Helper Functions for AnimalRankings Tests

// Utility function to print the AnimalRanking slice for debugging.
func printAnimalRankings(rankings []*AnimalRanking) string {
	var output string
	for _, r := range rankings {
		output += fmt.Sprintf("{ID: %d, Rank: %d, Name: %s, CreatedAt: %s}\n", r.ID, r.Rank, r.Name, r.CreatedAt)
	}
	return output
}

// Utility function to compare two slices of *AnimalRanking.
func equalAnimalRankingSlices(a, b []*AnimalRanking) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !equalAnimalRankings(a[i], b[i]) {
			return false
		}
	}
	return true
}

// Utility function to compare two *AnimalRanking instances.
func equalAnimalRankings(a, b *AnimalRanking) bool {
	return a.ID == b.ID && a.Rank == b.Rank && a.Name == b.Name && a.CreatedAt.Equal(b.CreatedAt)
}

// Initialize an in-memory SQLite database for testing the animal_rankings table.
func initAnimalRankingsTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// Create the animal_rankings table schema
	_, err = db.Exec(`
		CREATE TABLE animal_rankings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rank INTEGER NOT NULL UNIQUE,
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
		Rank      int
		Name      string
		CreatedAt time.Time
	}{
		{1, "Lion", parseTime("2024-09-25T10:00:00Z")},
		{2, "Tiger", parseTime("2024-09-25T10:05:00Z")},
		{3, "Elephant", parseTime("2024-09-25T10:10:00Z")},
		{4, "Leopard", parseTime("2024-09-25T10:15:00Z")},
		{5, "Wolf", parseTime("2024-09-25T10:20:00Z")},
	}

	for _, data := range sampleData {
		_, err := db.Exec(`INSERT INTO animal_rankings (rank, name, created_at) VALUES (?, ?, ?)`, data.Rank, data.Name, data.CreatedAt)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}
