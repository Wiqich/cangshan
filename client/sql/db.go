package sql

import (
	gosql "database/sql"
	"fmt"
	"github.com/yangchenxing/cangshan/application"
	"regexp"
)

func init() {
	application.RegisterModuleCreater("SQLDB", func() interface{} { return new(DB) })
}

var (
	lineSeperator = regexp.MustCompile("[\n\t ]+")
)

func normalizeSQLQuery(query string) string {
	return lineSeperator.ReplaceAllString(query, " ")
}

type DB struct {
	*gosql.DB
	Driver     string
	DataSource string
	Debug      string
}

func (db *DB) Initialize() error {
	var err error
	if db.DB, err = gosql.Open(db.Driver, db.DataSource); err != nil {
		return fmt.Errorf("open sql db fail: %s", err.Error())
	}
}

func (db *DB) Begin() (*Tx, error) {
	if db.Debug {
		logging.Debug("Begin SQL Transaction: %s", db.name)
	}
	if tx, err := db.DB.Begin(); err != nil {
		return nil, err
	} else {
		return &Tx{tx, db}, nil
	}
}

func (db *DB) Exec(query string, args ...interface{}) (gosql.Result, error) {
	if db.Debug {
		logging.Debug("SQL: query=\"%s\", params=%v", normalizeSQLQuery(query), args)
	}
	return db.DB.Exec(query, args...)
}

func (db *DB) Prepare(query string) (*Stmt, error) {
	if s, err := db.DB.Prepare(query); err != nil {
		return nil, err
	} else {
		return &Stmt{s, query, db}, nil
	}
}

func (db *DB) Query(query string, args ...interface{}) (*gosql.Rows, error) {
	if db.Debug {
		logging.Debug("SQL: query=\"%s\", params=%v", normalizeSQLQuery(query), args)
	}
	return db.DB.Query(query, args...)
}

func (db *DB) QueryRow(query string, args ...interface{}) *gosql.Row {
	if db.Debug {
		logging.Debug("SQL: query=\"%s\", params=%v", normalizeSQLQuery(query), args)
	}
	return db.DB.QueryRow(query, args...)
}
