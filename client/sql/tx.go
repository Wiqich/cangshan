package sql

import (
	gosql "database/sql"

	"github.com/yangchenxing/cangshan/logging"
)

// Tx is a wrapper of standard sql.Tx. It output query to debug log.
type Tx struct {
	*gosql.Tx
	db *DB
}

// Commit a transaction
func (tx *Tx) Commit() error {
	if tx.db.Debug {
		logging.Debug("Commit SQL Transaction")
	}
	return tx.Tx.Commit()
}

// Exec execute non-select query
func (tx Tx) Exec(query string, args ...interface{}) (gosql.Result, error) {
	if tx.db.Debug {
		logging.Debug("SQL: %s; %v", normalizeSQLQuery(query), args)
	}
	return tx.Tx.Exec(query, args...)
}

// Prepare a statement
func (tx Tx) Prepare(query string) (*Stmt, error) {
	stmt, err := tx.Tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &Stmt{stmt, query, tx.db}, nil
}

// Query multiple rows
func (tx Tx) Query(query string, args ...interface{}) (*gosql.Rows, error) {
	if tx.db.Debug {
		logging.Debug("SQL: %s; %v", normalizeSQLQuery(query), args)
	}
	return tx.Tx.Query(query, args...)
}

// QueryRow query single row
func (tx Tx) QueryRow(query string, args ...interface{}) *gosql.Row {
	if tx.db.Debug {
		logging.Debug("SQL: %s; %v", normalizeSQLQuery(query), args)
	}
	return tx.Tx.QueryRow(query, args...)
}

// Stmt transforms a non-transaction statement to a transaction statement
func (tx Tx) Stmt(stmt *Stmt) *Stmt {
	return &Stmt{tx.Tx.Stmt(stmt.Stmt), stmt.query, tx.db}
}

// Rollback a transaction
func (tx *Tx) Rollback() error {
	if tx.db.Debug {
		logging.Debug("Rollback SQL Transaction")
	}
	return tx.Tx.Rollback()
}
