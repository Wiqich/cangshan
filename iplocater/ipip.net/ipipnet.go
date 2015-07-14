package ipipnet

import (
	"fmt"
	"github.com/yangchenxing/cangshan/iplocater"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type record struct {
	start    uint32
	end      uint32
	country  *iplocater.Location
	province *iplocater.Location
	city     *iplocater.Location
	isp      []*iplocater.ISP
}

type IPData struct {
	VersionURL  string
	DownloadURL string
	Path        string
	version     string
}

func (ipdata *IPData) update() error {
	if _, err := os.Stat(name); err != nil {
		if err := os.MkdirAll(ipdata.Path, 0755); err != nil {
			return fmt.Errorf("make ipdata fold %s fail: %s", ipdata.Path, err.Error())
		}
	}
	var version string
	if resp, err := http.Get(ipdata.VersionURL); err != nil {
		return fmt.Errorf("check version fail: %s", err.Error())
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("check version fail: status=%d", resp.Status)
	} else if versionContent, err := ioutil.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("check version fail: %s", err.Error())
	} else {
		version = string(versionContent)
	}
	if version == ipdata.version {
		return nil
	}
	dataPath := filepath.Join(ipdata.Path, "data")
	if resp, err := http.Get(ipdata.DownloadURL); err != nil {
		return fmt.Errorf("download data fail: %s", err.Error())
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("download data fail: status=%d", resp.StatusCode)
	} else if data, err := ioutil.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("download data fail: %s", err.Error())
	} else if err := ioutil.WriteFile(dataPath+"."+version, data, 0755); err != nil {
		return fmt.Errorf("download data fail: %s", err.Error())
	} else if err := os.Rename(dataPath+"."+version, dataPath); err != nil {
		return fmt.Errorf("download data fail: %s", err.Error())
	}
	versionPath := filepath.Join(ipdata.Path, "version")
	if err := ioutil.WriteFile(versionPath+".tmp", []byte(version), 0755); err != nil {
		return fmt.Errorf("update version fail: %s", err.Error())
	} else if err := os.Rename(versionPath+".tmp", versionPath); err != nil {
		return fmt.Errorf("update version fail: %s", err.Error())
	}
	return ipdata.reload()
}

func (ipdata *IPData) reload() error {
	if data, err := ioutil.ReadFile(filepath.Join(ipdata.Path, "data")); err != nil {
		return fmt.Errorf("read data fail: %s", err.Error())
	} else {
		textOffset := binary.BigEndian.Uint32(data[:4]) - 1024
		ipTable.Rows = make([]IPRow, (textOffset-4-1024)/8)
		startIP := uint32(0)
		for i, offset := 0, uint32(1028); offset < textOffset; i, offset = i+1, offset+8 {
			ipTable.Rows[i].Start = startIP
			ipTable.Rows[i].End = binary.BigEndian.Uint32(data[offset : offset+4])
			dataOffset := textOffset + (uint32(data[offset+4]) | uint32(data[offset+5])<<8 | uint32(data[offset+6])<<16)
			dataLength := uint32(data[offset+7])
			ipTable.Rows[i].Data = data[dataOffset : dataOffset+dataLength]
			startIP = ipTable.Rows[i].End + 1
		}
	}
	return
}
