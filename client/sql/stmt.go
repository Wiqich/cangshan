package sql

import (
	gosql "database/sql"

	"github.com/yangchenxing/cangshan/logging"
)

// Stmt is a wrapper of standard sql.Stmt. It output query to debug log.
type Stmt struct {
	*gosql.Stmt
	query string
	db    *DB
}

// Exec executes a non-select query
func (s *Stmt) Exec(args ...interface{}) (Result, error) {
	if s.db.Debug {
		logging.Debug("SQL: %s; %v", s.query, args)
	}
	return s.Stmt.Exec(args...)
}

// Query multiple rows
func (s *Stmt) Query(args ...interface{}) (*Rows, error) {
	if s.db.Debug {
		logging.Debug("SQL: %s, %v", s.query, args)
	}
	rows, err := s.Stmt.Query(args...)
	return &Rows{rows}, err
}

// QueryRow query single row
func (s *Stmt) QueryRow(args ...interface{}) *Row {
	if s.db.Debug {
		logging.Debug("SQL: %s, %v", s.query, args)
	}
	return &Row{s.Stmt.QueryRow(args...)}
}
