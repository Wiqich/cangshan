package sql

import (
	gosql "database/sql"
)

type Stmt struct {
	*gosql.Stmt
	query string
	db    *DB
}

func (s *Stmt) Exec(args ...interface{}) (gosql.Result, error) {
	if s.db.Debug {
		logging.Debug("SQL: db=%s, query=\"%s\", params=%v",
			s.db.name, s.query, args)
	}
	return s.Stmt.Exec(args...)
}

func (s *Stmt) Query(args ...interface{}) (*gosql.Rows, error) {
	if s.db.Debug {
		logging.Debug("SQL: db=%s, query=\"%s\", params=%v",
			s.db.name, s.query, args)
	}
	return s.Stmt.Query(args...)
}

func (s *Stmt) QueryRow(args ...interface{}) *Row {
	if s.db.Debug {
		logging.Debug("SQL: db=%s, query=\"%s\", params=%v",
			s.db.name, s.query, args)
	}
	return s.Stmt.QueryRow(args...)
}
