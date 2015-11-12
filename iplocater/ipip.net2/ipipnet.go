package ipipnet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
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
	*ipipNetDataSet
	VersionURL      string
	DownloadURL     string
	Path            string
	UpdateInterval  time.Duration
	ForceInitialize bool
	DataSetCapacity int
	idDict          *iplocater.IDDict
	sections        *indexedSections
}

func (client *IPIPNet) Initialize() error {
	var err error
	iplocater.Debug("initialize ipip.net client")
	client.sections = &indexedSections{
		sections: make([]locationSection, 0),
		index:    make([]section, 256),
	}
	if client.ipipNetDataSet, err = loadDataSet(client.Path, client.DataSetCapacity, client.VersionURL, client.DownloadURL); err != nil {
		return fmt.Errorf("load ipip.net local data set %s fail: %s", client.Path, err.Error())
	}
	if client.idDict, err = iplocater.LoadIDDict(filepath.Join(client.Path, "iddict.csv")); err != nil {
		return fmt.Errorf("load iplocater id dictionary fail: %s", err.Error())
	}
	if err := client.load(); err != nil {
		return fmt.Errorf("load ipip.net data fail: %s", err.Error())
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

func (client *IPIPNet) update() error {
	if version, err := client.ipipNetDataSet.update(); err != nil {
		return fmt.Errorf("update ipip.net data set fail: %s", err.Error())
	} else if version == nil {
		iplocater.Debug("ipip.net data is up-to-date: %s", client.getVersion().String())
		return nil
	} else if err := client.load(); err != nil {
		return fmt.Errorf("load ipip.net data fail: %s", err.Error())
	} else {
		iplocater.Debug("update ipip.net data success: %s", client.getVersion().String())
	}
	return nil
}

func (client *IPIPNet) load() error {
	version := client.getVersion()
	if version == nil {
		if client.ForceInitialize {
			return client.update()
		} else {
			return errors.New("no data available")
		}
	}
	iplocater.Debug("load ipip.net data")
	if data, err := ioutil.ReadFile(version.getDataPath()); err != nil {
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
	iplocater.Debug("load ipip.net data success: %s", version.String())
	return nil
}

func (client *IPIPNet) parseLocation(data []byte) *iplocater.Location {
	fields := strings.Split(string(data), "\t")
	return client.idDict.GetLocation(fields[:3], strings.Split(fields[len(fields)-1], "/"))
}
