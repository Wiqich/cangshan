package sql

import (
	"errors"
	"fmt"
	"github.com/yangchenxing/cangshan/client/sql"
	"github.com/yangchenxing/cangshan/logging"
	"time"
)

type SQLKV struct {
	DB            *sql.DB
	CreateQueries   []string
	GetQuery      string
	SetQuery      string
	CleanQuery    string
	CleanInterval time.Duration
}

func (kv *SQLKV) Initialize() error {
	if kv.CreateQuery != "" {
        for _, query := range kv.CreateQueries {
            if _, err := kv.DB.Exec(kv.query); err != nil {
                return fmt.Errorf("Create SQLKV table fail: %s", err.Error())
            }
        }
	}
	if kv.GetQuery == "" {
		return errors.New("Missing GetQuery")
	} else if kv.SetQuery == "" {
		return errors.New("Missing SetQuery")
	}
	if kv.CleanQuery != "" {
		go kv.autoClean()
	}
	return nil
}

func (kv *SQLKV) Ping() error {
	return kv.DB.Ping()
}

func (kv *SQLKV) Get(key string) ([]byte, error) {
	row := kv.DB.QueryRow(kv.GetQuery)
	var value []byte
	if err := row.Scan(&value); err != nil {
		return nil, err
	} else {
		return value, nil
	}
}

func (kv *SQLKV) Set(key string, value []byte, maxAge time.Duration) error {
    _, err := kv.DB.Exec(kv.SetQuery, key, value, time.Now().Add(maxAge).Unix())
    return err
}

func (kv *SQLKV) autoClean() {
	for {
		if _, err := kv.DB.Exec(CleanQuery, time.Now().Unix(); err != nil {
			logging.Error("Clean SQLKV fail: %s", err.Error())
		}
		time.Sleep(kv.CleanInterval)
	}
}
