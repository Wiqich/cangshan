package kvlite

import (
	"database/sql"
	"fmt"
	"hash"
	"hash/crc64"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	CREATE_QUERY = `
CREATE TABLE IF NOT EXISTS data(
	id UNSIGNED INTEGER PRIMARY KEY,
	key TEXT,
	value TEXT,
	expire UNSIGNED INTEGER
)`
	SEARCH_QUERY = `
SELECT key, value, expire
FROM   data
WHERE  id=?`
	INSERT_QUERY = `
INSERT OR REPLACE INTO data(id, key, value, expire) VALUES(?, ?, ?, ?)`
	DELETE_QUERY = `
DELETE FROM data WHERE id=?`
	CLEAN_QUERY = `
DELETE FROM data WHERE expire>0 AND expire<=?`
)

type HashMethodFactory func() hash.Hash64

type KVLite struct {
	db      *sql.DB
	factory HashMethodFactory
}

func NewKVLite(path string, factory HashMethodFactory) (*KVLite, error) {
	if db, err := sql.Open("sqlite3", path); err != nil {
		return nil, fmt.Errorf("create database fail: %s", err.Error())
	} else if _, err := db.Exec(CREATE_QUERY); err != nil {
		return nil, fmt.Errorf("create table fail: %s", err.Error())
	} else if _, err := db.Exec(CLEAN_QUERY, time.Now().Unix()); err != nil {
		return nil, fmt.Errorf("clean expired data fail: %s", err.Error())
	} else {
		if factory == nil {
			factory = func() hash.Hash64 {
				return crc64.New(crc64.MakeTable(crc64.ISO))
			}
		}
		return &KVLite{
			db:      db,
			factory: factory,
		}, nil
	}
}

func (lite KVLite) hash(key string) int64 {
	hashMethod := lite.factory()
	hashMethod.Write([]byte(key))
	return int64(hashMethod.Sum64())
}

func (lite KVLite) Get(key string) (string, error) {
	hashKey := lite.hash(key)
	row := lite.db.QueryRow(SEARCH_QUERY, hashKey)
	var itemKey, itemValue string
	var expire int64
	if err := row.Scan(&itemKey, &itemValue, &expire); err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("sqlite error: query=%s, error=%s", SEARCH_QUERY, err.Error())
	}
	if itemKey != key {
		return "", nil
	}
	if expire > 0 && expire < time.Now().Unix() {
		lite.Delete(key)
		return "", nil
	}
	return itemValue, nil
}

func (lite KVLite) Set(key, value string, maxage time.Duration) error {
	hashKey := lite.hash(key)
	expire := int64(0)
	if maxage > 0 {
		expire = time.Now().Add(maxage).Unix()
	}
	if _, err := lite.db.Exec(INSERT_QUERY, hashKey, key, value, expire); err != nil {
		return err
	}
	return nil
}

func (lite KVLite) Delete(key string) error {
	hashKey := lite.hash(key)
	if _, err := lite.db.Exec(DELETE_QUERY, hashKey); err != nil {
		return err
	}
	return nil
}
