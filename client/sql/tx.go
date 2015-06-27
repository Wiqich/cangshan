package sql

import (
	gosql "database/sql"
	"logging"
)

type Tx struct {
	*gosql.Tx
	db *DB
}

func (tx *Tx) Commit() error {
	if tx.db.Debug {
		logging.Debug("Commit SQL Transaction: %s", tx.db.name)
	}
	return tx.Tx.Commit()
}

func (tx Tx) Exec(query string, args ...interface{}) (gosql.Result, error) {
	if tx.db.Debug {
		logging.Debug("SQL: db=%s, query=\"%s\", params=%v",
			tx.db.name, normalizeSQLQuery(query), args)
	}
	return tx.Tx.Exec(query, args...)
}

func (tx Tx) Prepare(query string) (*Stmt, error) {
	if stmt, err := tx.Tx.Prepare(query); err != nil {
		return nil, err
	} else {
		return &Stmt{stmt, tx.db}, nil
	}
}

func (tx Tx) Query(query string, args ...interface{}) (*gosql.Rows, error) {
	if tx.db.Debug {
		logging.Debug("SQL: db=%s, query=\"%s\", params=%v",
			tx.db.name, normalizeSQLQuery(query), args)
	}
	return tx.Tx.Query(query, args...)
}

func (tx Tx) QueryRow(query string, args ...interface{}) *gosql.Row {
	if tx.db.Debug {
		logging.Debug("SQL: db=%s, query=\"%s\", params=%v",
			tx.db.name, normalizeSQLQuery(query), args)
	}
	return tx.Tx.QueryRow(query, args...)
}

func (tx Tx) Stmt(stmt *Stmt) *Stmt {
	return &Stmt{tx.Tx.Stmt(stmt.Stmt), stmt.query, tx.db}
}

func (tx *Tx) Rollback() error {
	if tx.db.Debug {
		logging.Debug("Rollback SQL Transaction: %s", tx.db.name)
	}
	return tx.Tx.Rollback()
}
