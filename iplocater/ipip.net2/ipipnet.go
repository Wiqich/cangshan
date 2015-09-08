package ipipnet

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/iplocater"
)

const (
	defaultUpdateInterval = time.Hour
)

func init() {
	application.RegisterModulePrototype("IPIPNet", new(IPIPNet))
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
	VersionURL      string
	DownloadURL     string
	Path            string
	UpdateInterval  time.Duration
	ForceInitialize bool
	idDict          *iplocater.IDDict
	version         string
	sections        *indexedSections
}

func (client *IPIPNet) Initialize() error {
	iplocater.Debug("initialize ipip.net client")
	if _, err := os.Stat(client.Path); err != nil {
		if err := os.MkdirAll(client.Path, 0755); err != nil {
			return fmt.Errorf("ensure data directory fail: %s", err.Error())
		}
	}
	var err error
	if client.idDict, err = iplocater.LoadIDDict(filepath.Join(client.Path, "iddict.csv")); err != nil {
		return fmt.Errorf("load iplocater id dictionary fail: %s", err.Error())
	} else {
		iplocater.Debug("load iplocater id dictionary success")
		text, _ := json.MarshalIndent(client.idDict, "", "    ")
		ioutil.WriteFile(filepath.Join(client.Path, "iddict.dump.json"), text, 0755)
	}
	if err := client.selectLastVersion(); err != nil {
		return err
	}
	if err := client.load(); err != nil {
		return err
	}
	if client.UpdateInterval > 0 {
		go func() {
			iplocater.Debug("ipip.net update interval: %s", client.UpdateInterval)
			for {
				if err := client.update(); err != nil {
					iplocater.Error("update ipip.net data fail: %s", err.Error())
				}
				time.Sleep(client.UpdateInterval)
			}
		}()
	}
	iplocater.Debug("initialize ipip.net client success")
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

func (client *IPIPNet) download() (content []byte, version string, err error) {
	var response *http.Response
	var etag string
	var timestamp time.Time
	if response, err = http.Get(client.DownloadURL); err != nil {
	} else if etag = response.Header.Get("ETag"); etag == "" {
		err = errors.New("missing ETag header")
	} else if !strings.HasPrefix(etag, "sha1-") {
		err = fmt.Errorf("unsupported ETag header: %s", etag)
	} else if ts := response.Header.Get("Last-Modified"); ts == "" {
		err = errors.New("missing Last-Modified header")
	} else if timestamp, err = time.Parse(time.RFC1123, ts); err != nil {
		err = fmt.Errorf("invalid Last-Modified: %s", ts)
	} else if content, err = ioutil.ReadAll(response.Body); err != nil {
		err = fmt.Errorf("read http response body fail: %s", err.Error())
	} else if sum := fmt.Sprintf("%x", sha1.Sum(content)); sum != etag[5:] {
		err = fmt.Errorf("check sum fail: actual=%s, expected=%s", sum, etag[5:])
	} else {
		version = fmt.Sprintf("%s_%s", timestamp.Format("20060102150405"), etag[5:])
		iplocater.Debug("download ipip.net data success")
	}
	return
}

func (client *IPIPNet) update() error {
	iplocater.Debug("update ipip.net data")
	if updateVersion, err := client.checkUpdate(); err != nil {
		// iplocater.Error("check ipip.net update fail: %s", err)
		return fmt.Errorf("check update fail: %s", err.Error())
	} else if updateVersion != nil {
		if filename, err := client.download(); err == nil {
			updateVersion.Filename = filename
			client.version = updateVersion
			client.version.save(filepath.Join(client.Path, ".version"))
			iplocater.Debug("download ipip.net data success")
			if err := client.load(); err != nil {
				// iplocater.Error("load ipip.net data fail: %s", err)
				return fmt.Errorf("load ipip.net data fail: %s", err)
			}
			iplocater.Debug("load new ipip.net data success")
		} else {
			// iplocater.Error("download ipip.net data fail: %s", err)
			return fmt.Errorf("download ipip.net data fail: %s", err)
		}
	} else {
		iplocater.Debug("ipip.net data is up-to-date")
	}
	return nil
}

func (client *IPIPNet) load() error {
	if client.version == "" {
		iplocater.Debug("no available version")
		return nil
	}
	iplocater.Debug("load ipip.net data")
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
	iplocater.Debug("load ipip.net data success")
	return nil
}

func (client *IPIPNet) parseLocation(data []byte) *iplocater.Location {
	fields := strings.Split(string(data), "\t")
	return client.idDict.GetLocation(fields[:3], strings.Split(fields[len(fields)-1], "/"))
}
