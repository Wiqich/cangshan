package sqlkv

import (
	"errors"
	"fmt"
	"github.com/yangchenxing/cangshan/client/sql"
	"github.com/yangchenxing/cangshan/client/kv"
	"github.com/yangchenxing/cangshan/logging"
	"time"
	"database/sql"
)

type SQLKV struct {
	DB            *sql.DB
	CreateQueries   []string
	GetQuery      string
	SetQuery      string
	CleanQuery    string
	CleanInterval time.Duration
}

func (k *SQLKV) Initialize() error {
	if k.CreateQuery != "" {
        for _, query := range k.CreateQueries {
            if _, err := k.DB.Exec(k.query); err != nil {
                return fmt.Errorf("Create SQLKV table fail: %s", err.Error())
            }
        }
	}
	if k.GetQuery == "" {
		return errors.New("Missing GetQuery")
	} else if k.SetQuery == "" {
		return errors.New("Missing SetQuery")
	}
	if k.CleanQuery != "" {
		go k.autoClean()
	}
	return nil
}

func (k *SQLKV) Ping() error {
	return k.DB.Ping()
}

func (k *SQLKV) Get(key string) ([]byte, error) {
	row := k.DB.QueryRow(k.GetQuery)
	var value []byte
	if err := row.Scan(&value); err == sql.ErrNoRows {
		return nil, kv.ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return value, nil
	}
}

func (k *SQLKV) Set(key string, value []byte, maxAge time.Duration) error {
    _, err := k.DB.Exec(k.SetQuery, key, value, time.Now().Add(maxAge).Unix())
    return err
}

func (k *SQLKV) autoClean() {
	for {
		if _, err := k.DB.Exec(CleanQuery, time.Now().Unix(); err != nil {
			logging.Error("Clean SQLKV fail: %s", err.Error())
		}
		time.Sleep(k.CleanInterval)
	}
}
