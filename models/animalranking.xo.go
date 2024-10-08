// Package models contains generated code for schema 'platform'.
package models

// Code generated by xo. DO NOT EDIT.

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// AnimalRanking represents a row from 'platform.animal_rankings'.
type AnimalRanking struct {
	ID        int       `json:"id"`         // id
	Rank      int       `json:"rank"`       // rank
	Name      string    `json:"name"`       // name
	CreatedAt time.Time `json:"created_at"` // created_at
	UpdatedAt time.Time `json:"updated_at"` // updated_at
	// xo fields
	_exists, _deleted bool
}

// Exists returns true when the [AnimalRanking] exists in the database.
func (ar *AnimalRanking) Exists() bool {
	return ar._exists
}

// Deleted returns true when the [AnimalRanking] has been marked for deletion
// from the database.
func (ar *AnimalRanking) Deleted() bool {
	return ar._deleted
}

// Insert inserts the [AnimalRanking] to the database.
func (ar *AnimalRanking) Insert(ctx context.Context, db DB) error {
	switch {
	case ar._exists: // already exists
		return logerror(&ErrInsertFailed{ErrAlreadyExists})
	case ar._deleted: // deleted
		return logerror(&ErrInsertFailed{ErrMarkedForDeletion})
	}
	// insert (primary key generated and returned by database)
	const sqlstr = `INSERT INTO platform.animal_rankings (` +
		`rank, name, created_at, updated_at` +
		`) VALUES (` +
		`?, ?, ?, ?` +
		`)`
	// run
	logf(sqlstr, ar.Rank, ar.Name, ar.CreatedAt, ar.UpdatedAt)
	res, err := db.ExecContext(ctx, sqlstr, ar.Rank, ar.Name, ar.CreatedAt, ar.UpdatedAt)
	if err != nil {
		return logerror(err)
	}
	// retrieve id
	id, err := res.LastInsertId()
	if err != nil {
		return logerror(err)
	} // set primary key
	ar.ID = int(id)
	// set exists
	ar._exists = true
	return nil
}

// Update updates a [AnimalRanking] in the database.
func (ar *AnimalRanking) Update(ctx context.Context, db DB) error {
	switch {
	case !ar._exists: // doesn't exist
		return logerror(&ErrUpdateFailed{ErrDoesNotExist})
	case ar._deleted: // deleted
		return logerror(&ErrUpdateFailed{ErrMarkedForDeletion})
	}
	// update with primary key
	const sqlstr = `UPDATE platform.animal_rankings SET ` +
		`rank = ?, name = ?, created_at = ?, updated_at = ? ` +
		`WHERE id = ?`
	// run
	logf(sqlstr, ar.Rank, ar.Name, ar.CreatedAt, ar.UpdatedAt, ar.ID)
	if _, err := db.ExecContext(ctx, sqlstr, ar.Rank, ar.Name, ar.CreatedAt, ar.UpdatedAt, ar.ID); err != nil {
		return logerror(err)
	}
	return nil
}

// Save saves the [AnimalRanking] to the database.
func (ar *AnimalRanking) Save(ctx context.Context, db DB) error {
	if ar.Exists() {
		return ar.Update(ctx, db)
	}
	return ar.Insert(ctx, db)
}

// Upsert performs an upsert for [AnimalRanking].
func (ar *AnimalRanking) Upsert(ctx context.Context, db DB) error {
	switch {
	case ar._deleted: // deleted
		return logerror(&ErrUpsertFailed{ErrMarkedForDeletion})
	}
	// upsert
	const sqlstr = `INSERT INTO platform.animal_rankings (` +
		`id, rank, name, created_at, updated_at` +
		`) VALUES (` +
		`?, ?, ?, ?, ?` +
		`)` +
		` ON DUPLICATE KEY UPDATE ` +
		`rank = VALUES(rank), name = VALUES(name), created_at = VALUES(created_at), updated_at = VALUES(updated_at)`
	// run
	logf(sqlstr, ar.ID, ar.Rank, ar.Name, ar.CreatedAt, ar.UpdatedAt)
	if _, err := db.ExecContext(ctx, sqlstr, ar.ID, ar.Rank, ar.Name, ar.CreatedAt, ar.UpdatedAt); err != nil {
		return logerror(err)
	}
	// set exists
	ar._exists = true
	return nil
}

// Delete deletes the [AnimalRanking] from the database.
func (ar *AnimalRanking) Delete(ctx context.Context, db DB) error {
	switch {
	case !ar._exists: // doesn't exist
		return nil
	case ar._deleted: // deleted
		return nil
	}
	// delete with single primary key
	const sqlstr = `DELETE FROM platform.animal_rankings ` +
		`WHERE id = ?`
	// run
	logf(sqlstr, ar.ID)
	if _, err := db.ExecContext(ctx, sqlstr, ar.ID); err != nil {
		return logerror(err)
	}
	// set deleted
	ar._deleted = true
	return nil
}

// AnimalRankingKeysetPage retrieves a page of [AnimalRanking] records using keyset pagination with dynamic filtering.
//
// The keyset pagination retrieves results after or before a specific value (`key`)
// for a given column (`column`) with a limit (`limit`) and order (`ASC` or `DESC`).
//
// If `order` is `ASC`, it retrieves records where the value of `column` is greater than `key`.
// If `order` is `DESC`, it retrieves records where the value of `column` is less than `key`.
//
// Filters are dynamically provided via a `filters` map, where keys are column names and values are either single values or slices for `IN` clauses.
func AnimalRankingKeysetPage(ctx context.Context, db DB, column string, key interface{}, limit int, order string, filters map[string]interface{}) ([]*AnimalRanking, *AnimalRanking, error) {
	if order != "ASC" && order != "DESC" {
		return nil, nil, fmt.Errorf("invalid order: %s", order)
	}

	// Start building the query
	query := fmt.Sprintf(
		`SELECT * FROM animal_rankings 
         WHERE %s %s ?`,
		column, condition(order), // Ensure this is returning a valid operator
	)

	// Arguments for the query
	args := []interface{}{key}

	// Dynamically add filters from the `filters` map to the query
	for field, value := range filters {
		switch v := value.(type) {
		case []int:
			if len(v) > 0 {
				placeholders := make([]string, len(v))
				for i := range v {
					placeholders[i] = "?"
					args = append(args, v[i])
				}
				query += fmt.Sprintf(" AND %s IN (%s)", field, strings.Join(placeholders, ", "))
			}
		case []string:
			if len(v) > 0 {
				placeholders := make([]string, len(v))
				for i := range v {
					placeholders[i] = "?"
					args = append(args, v[i])
				}
				query += fmt.Sprintf(" AND %s IN (%s)", field, strings.Join(placeholders, ", "))
			}
		default:
			query += fmt.Sprintf(" AND %s = ?", field)
			args = append(args, value)
		}
	}

	// Finalize the query with the order and limit
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT ?", column, order)
	args = append(args, limit)

	// Log the final query for debugging purposes
	log.Printf("Executing query: %s with args: %v", query, args)

	// Execute the query
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, logerror(err)
	}
	defer rows.Close()

	var results []*AnimalRanking
	var lastItem *AnimalRanking // Variable to store the last item

	for rows.Next() {
		ar := AnimalRanking{
			_exists: true,
		}
		if err := rows.Scan(
			&ar.ID, &ar.Rank, &ar.Name, &ar.CreatedAt, &ar.UpdatedAt,
		); err != nil {
			return nil, nil, logerror(err)
		}
		results = append(results, &ar)
	}

	// Check for errors during row iteration.
	if err := rows.Err(); err != nil {
		return nil, nil, logerror(err)
	}

	// If we have results, set the lastItem to the last element in results.
	if len(results) > 0 {
		lastItem = results[len(results)-1]
	}

	return results, lastItem, nil
}

// AnimalRankingByID retrieves a row from 'platform.animal_rankings' as a [AnimalRanking].
//
// Generated from index 'animal_rankings_id_pkey'.
func AnimalRankingByID(ctx context.Context, db DB, id int) (*AnimalRanking, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, rank, name, created_at, updated_at ` +
		`FROM platform.animal_rankings ` +
		`WHERE id = ?`
	// run
	logf(sqlstr, id)
	ar := AnimalRanking{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, id).Scan(&ar.ID, &ar.Rank, &ar.Name, &ar.CreatedAt, &ar.UpdatedAt); err != nil {
		return nil, logerror(err)
	}
	return &ar, nil
}

// AnimalRankingByRank retrieves a row from 'platform.animal_rankings' as a [AnimalRanking].
//
// Generated from index 'rank'.
func AnimalRankingByRank(ctx context.Context, db DB, rank int) (*AnimalRanking, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, rank, name, created_at, updated_at ` +
		`FROM platform.animal_rankings ` +
		`WHERE rank = ?`
	// run
	logf(sqlstr, rank)
	ar := AnimalRanking{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, rank).Scan(&ar.ID, &ar.Rank, &ar.Name, &ar.CreatedAt, &ar.UpdatedAt); err != nil {
		return nil, logerror(err)
	}
	return &ar, nil
}
