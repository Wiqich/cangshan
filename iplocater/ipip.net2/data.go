package ipipnet

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yangchenxing/cangshan/iplocater"
)

type ipipNetData struct {
	sync.Mutex
	version      string
	etag         string
	lastModified time.Time
	path         string
}

func (data ipipNetData) String() string {
	return fmt.Sprintf("IPIPNetData{version=%s, etag=%s, lastModifed=%s}",
		data.version, data.etag, data.lastModified.Format(time.RFC3339))
}

func (data ipipNetData) getDataPath() string {
	return filepath.Join(data.path, "data")
}

func (data ipipNetData) getMetadataPath() string {
	return filepath.Join(data.path, ".version")
}

func (data *ipipNetData) download(url string) error {
	data.Lock()
	defer data.Unlock()
	if response, err := http.Get(url); err != nil {
		return fmt.Errorf("download %s fail: %s", url, err.Error())
	} else if etag := response.Header.Get("ETag"); etag == "" {
		return errors.New("missing ETag header")
	} else if !strings.HasPrefix(etag, "sha1-") {
		return fmt.Errorf("unsupported ETag header: %s", etag)
	} else if ts := response.Header.Get("Last-Modified"); ts == "" {
		return errors.New("missing Last-Modified header")
	} else if timestamp, err := time.Parse(time.RFC1123, ts); err != nil {
		return fmt.Errorf("invalid Last-Modified: %s", ts)
	} else if content, err := ioutil.ReadAll(response.Body); err != nil {
		return fmt.Errorf("read http response body fail: %s", err.Error())
	} else if sum := fmt.Sprintf("%x", sha1.Sum(content)); sum != etag[5:] {
		return fmt.Errorf("check sum fail: actual=%s, expected=%s", sum, etag[5:])
	} else if err := ioutil.WriteFile(data.getDataPath(), content, 0755); err != nil {
		return fmt.Errorf("write %s fail: %s", data.getDataPath(), err.Error())
	} else {
		data.lastModified = timestamp
		data.etag = etag
		if err := data.save(); err != nil {
			return err
		}
		iplocater.Debug("download ipip.net data success: %s", data.String())
		return nil
	}
}

func (data *ipipNetData) update(url string) (bool, error) {
	if data == nil {
		return false, nil
	} else if response, err := http.Head(url); err != nil {
		return false, fmt.Errorf("get head of %s fail: %s", url, err.Error())
	} else if etag := response.Header.Get("ETag"); etag == "" {
		return false, errors.New("missing ETag header")
	} else if !strings.HasPrefix(etag, "sha1-") {
		return false, fmt.Errorf("unsupported ETag header: %s", etag)
	} else if ts := response.Header.Get("Last-Modified"); ts == "" {
		return false, errors.New("missing Last-Modified header")
	} else if timestamp, err := time.Parse(time.RFC1123, ts); err != nil {
		return false, fmt.Errorf("invalid Last-Modified: %s", ts)
	} else if etag != data.etag {
		if err := data.download(url); err != nil {
			return false, err
		}
		iplocater.Warn("ipip.net data %s updated: %s", data.version, data.String())
		return true, nil
	} else if !timestamp.Equal(data.lastModified) {
		iplocater.Warn("ipip.net data %s got new timestamp (%s vs. %s) without new data",
			data.version, timestamp.Format(time.RFC1123), data.lastModified.Format(time.RFC1123))
		return false, nil
	}
	return false, nil
}

func loadRemoteData(path, version, url string) (*ipipNetData, error) {
	if err := os.Mkdir(path, 0755); err != nil {
		return nil, fmt.Errorf("create ipip.net data directory %s fail: %s", path, err.Error())
	}
	data := &ipipNetData{
		version: version,
		path:    path,
	}
	if err := data.download(url); err != nil {
		return nil, fmt.Errorf("download ipip.net data fail: %s", err.Error())
	}
	return data, nil
}

func loadLocalData(path string) (*ipipNetData, error) {
	data := &ipipNetData{
		path: path,
	}
	metadata := make(map[string]interface{})
	if content, err := ioutil.ReadFile(data.getMetadataPath()); err != nil {
		return nil, fmt.Errorf("read metadata %s fail: %s", data.getMetadataPath(), err.Error())
	} else if err := json.Unmarshal(content, &metadata); err != nil {
		return nil, fmt.Errorf("unmarshal metadata from %s fail: %s", data.getMetadataPath(), err.Error())
	}
	var ok bool
	var err error
	if data.version, ok = metadata["version"].(string); !ok {
		return nil, fmt.Errorf("no version in metadata %s", data.getMetadataPath())
	} else if data.etag, ok = metadata["etag"].(string); !ok {
		return nil, fmt.Errorf("no etag in metadata %s", data.getMetadataPath())
	} else if lastModified, ok := metadata["last_modified"].(string); !ok {
		return nil, fmt.Errorf("no last_modified in metadata %s", data.getMetadataPath())
	} else if data.lastModified, err = time.Parse(time.RFC1123, lastModified); err != nil {
		return nil, fmt.Errorf("parse last_modified in metadata %s fail: %s", data.getMetadataPath(), err.Error())
	}
	return data, nil
}

func (data *ipipNetData) save() error {
	metadata := map[string]interface{}{
		"version":       data.version,
		"etag":          data.etag,
		"last_modified": data.lastModified.Format(time.RFC1123),
	}
	if content, err := json.Marshal(metadata); err != nil {
		return fmt.Errorf("marshal metadata fail: %s", err.Error())
	} else if err := ioutil.WriteFile(data.getMetadataPath(), content, 0755); err != nil {
		return fmt.Errorf("save metadata to %s fail: %s", data.getMetadataPath(), err.Error())
	}
	return nil
}

func (data *ipipNetData) remove() error {
	return os.Remove(data.path)
}
