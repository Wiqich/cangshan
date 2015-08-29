package ipipnet

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/iplocater"
	"github.com/yangchenxing/cangshan/logging"
)

func init() {
	application.RegisterModulePrototype("IPIPNet", new(IPIPNet))
}

type IPIPNetDataVersion struct {
	Filename string `json:"filename"`
	Version  string `json:"version"`
}

func (version *IPIPNetDataVersion) load(path string) error {
	if content, err := ioutil.ReadFile(path); err != nil {
		return fmt.Errorf("read meta file fail: %s", err.Error())
	} else if err := json.Unmarshal(content, version); err != nil {
		return fmt.Errorf("unmarshal meta file fail: %s", err.Error())
	}
	return nil
}

func (version *IPIPNetDataVersion) save(path string) error {
	if content, err := json.Marshal(version); err != nil {
		return fmt.Errorf("marshal meta fail: %s", err.Error())
	} else if err := ioutil.WriteFile(path, content, 0755); err != nil {
		return fmt.Errorf("write meta fail: %s", err.Error())
	}
	return nil
}

type section struct {
	start uint32
	end   uint32
}

type locationSection struct {
	*iplocater.Location
	section
}

type indexedSections struct {
	sections []locationSection
	index    []section
}

type IPIPNet struct {
	VersionURL     string
	DownloadURL    string
	Path           string
	UpdateInterval time.Duration
	idDict         *iplocater.IDDict
	version        *IPIPNetDataVersion
	sections       *indexedSections
}

func (client *IPIPNet) Initialize() error {
	logging.Debug("initialize ipip.net client")
	if _, err := os.Stat(client.Path); err != nil {
		if err := os.MkdirAll(client.Path, 0755); err != nil {
			return fmt.Errorf("ensure data directory fail: %s", err.Error())
		}
	}
	var err error
	if client.idDict, err = iplocater.LoadIDDict(filepath.Join(client.Path, "iddict.csv")); err != nil {
		return fmt.Errorf("load iplocater id dictionary fail: %s", err.Error())
	} else {
		logging.Debug("load iplocater id dictionary success")
		text, _ := json.MarshalIndent(client.idDict, "", "    ")
		ioutil.WriteFile(filepath.Join(client.Path, "iddict.dump.json"), text, 0755)
	}
	client.version = new(IPIPNetDataVersion)
	client.version.load(filepath.Join(client.Path, ".version"))
	logging.Debug("ipip.net data version: %v", client.version)
	if err := client.update(); err != nil {
		return err
	}
	if client.UpdateInterval > 0 {
		go func() {
			logging.Debug("ipip.net update interval: %s", client.UpdateInterval)
			for {
				time.Sleep(client.UpdateInterval)
				if err := client.update(); err != nil {
					logging.Error("update ipip.net data fail: %s", err.Error())
				}
			}
		}()
	}
	logging.Debug("initialize ipip.net client success")
	// logging.Debug("section size: %d, index size: %d", len(client.sections.sections), len(client.sections.index))
	return nil
}

func (client *IPIPNet) Locate(ip net.IP) (*iplocater.Location, error) {
	if ip = ip.To4(); ip == nil {
		return nil, errors.New("Not a IPv4 address")
	}
	v := binary.BigEndian.Uint32([]byte(ip))
	idx := client.sections.index[v>>24]
	for start, end := int(idx.start), int(idx.end); start <= end; {
		middle := (start + end) / 2
		section := client.sections.sections[middle]
		if v < section.start {
			end = middle - 1
		} else if v > section.end {
			start = middle + 1
		} else {
			return section.Location, nil
		}
	}
	return nil, nil
}

func (client *IPIPNet) checkUpdate() (*IPIPNetDataVersion, error) {
	if resp, err := http.Get(client.VersionURL); err != nil {
		return nil, fmt.Errorf("check version fail: %s", err.Error())
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("check version fail: status=%d", resp.Status)
	} else if version, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, fmt.Errorf("check version fail: %s", err.Error())
	} else if client.version.Version != string(version) {
		logging.Debug("found ipip.net data new version: %s", string(version))
		return &IPIPNetDataVersion{
			Version: string(version),
		}, nil
	} else {
		return nil, nil
	}
}

func (client *IPIPNet) download() (string, error) {
	response, err := http.Get(client.DownloadURL)
	if err != nil {
		return "", err
	}
	var filename string
	if contentDisposition := response.Header.Get("Content-Disposition"); contentDisposition == "" {
		return "", errors.New("missing Content-Disposition header")
	} else if _, contentParams, err := mime.ParseMediaType(contentDisposition); err != nil {
		return "", fmt.Errorf("parse Content-Disposition header fail: %s", err.Error())
	} else if filename = contentParams["filename"]; filename == "" {
		return "", errors.New("no filename in Content-Disposition header")
	}
	var checksum string
	if etag := response.Header.Get("ETag"); etag == "" {
		return "", errors.New("missing ETag header")
	} else if !strings.HasPrefix(etag, "sha1-") {
		return "", fmt.Errorf("unsupported ETag header: %s", etag)
	} else {
		checksum = etag[5:]
	}
	if content, err := ioutil.ReadAll(response.Body); err != nil {
		return "", fmt.Errorf("read http response body fail: %s", err.Error())
	} else {
		sum := sha1.Sum(content)
		if sum := hex.EncodeToString(sum[:]); sum != checksum {
			return "", fmt.Errorf("check sum fail: actual=%s, expected=%s", sum, checksum)
		} else if err := ioutil.WriteFile(filepath.Join(client.Path, filename), content, 0755); err != nil {
			return "", fmt.Errorf("write data file fail: %s", err.Error())
		}
	}
	logging.Debug("download ipip.net data success")
	return filename, nil
}

func (client *IPIPNet) update() error {
	logging.Debug("update ipip.net data")
	needToLoad := client.sections == nil
	if updateVersion, err := client.checkUpdate(); err != nil {
		return fmt.Errorf("check update fail: %s", err.Error())
	} else if updateVersion != nil {
		if filename, err := client.download(); err == nil {
			updateVersion.Filename = filename
			client.version = updateVersion
			client.version.save(filepath.Join(client.Path, ".version"))
			logging.Debug("download ipip.net data success")
			needToLoad = true
		} else {
			return fmt.Errorf("download ipip.net data file fail: %s", err.Error())
		}
	} else {
		logging.Debug("ipip.net data is up-to-date")
	}
	if needToLoad {
		if err := client.load(); err != nil {
			return fmt.Errorf("load ipip.net data fail: %s", err.Error())
		}
	}
	return nil
}

func (client *IPIPNet) load() error {
	logging.Debug("load ipip.net data")
	if data, err := ioutil.ReadFile(filepath.Join(client.Path, client.version.Filename)); err != nil {
		return fmt.Errorf("read data fail: %s", err.Error())
	} else {
		textOffset := binary.BigEndian.Uint32(data[:4]) - 1024
		idxSections := &indexedSections{
			sections: make([]locationSection, (textOffset-4-1024)/8),
			index:    make([]section, 256),
		}
		startIP := uint32(1)
		for i, offset := uint32(0), uint32(1028); offset < textOffset; i, offset = i+1, offset+8 {
			endIP := binary.BigEndian.Uint32(data[offset : offset+4])
			idxSections.sections[i].start = startIP
			idxSections.sections[i].end = endIP
			dataOffset := textOffset + (uint32(data[offset+4]) | uint32(data[offset+5])<<8 | uint32(data[offset+6])<<16)
			dataLength := uint32(data[offset+7])
			idxSections.sections[i].Location = client.parseLocation(data[dataOffset : dataOffset+dataLength])
			startIP = endIP + 1
			idxSections.index[endIP>>24].end = i
		}
		for i, _ := range idxSections.index {
			if i == 0 {
				idxSections.index[i].start = uint32(0)
			} else {
				idxSections.index[i].start = idxSections.index[i-1].end + 1
			}
		}
		client.sections = idxSections
	}
	logging.Debug("load ipip.net data success")
	return nil
}

func (client *IPIPNet) parseLocation(data []byte) *iplocater.Location {
	fields := strings.Split(string(data), "\t")
	return client.idDict.GetLocation(fields[:3], strings.Split(fields[len(fields)-1], "/"))
}
