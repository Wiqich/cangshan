package ipipnet

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sort"

	"github.com/yangchenxing/cangshan/iplocater"
)

type ipipNetDataSet struct {
	capacity    int
	path        string
	datas       []*ipipNetData
	versionURL  string
	downloadURL string
}

func loadDataSet(path string, capacity int, versionURL, downloadURL string) (*ipipNetDataSet, error) {
	dataset := &ipipNetDataSet{
		capacity:    capacity,
		path:        path,
		versionURL:  versionURL,
		downloadURL: downloadURL,
		datas:       make([]*ipipNetData, 0, capacity+1),
	}
	if infos, err := ioutil.ReadDir(path); err != nil {
		return nil, fmt.Errorf("read ipip.net data set directory %s fail: %s",
			path, err.Error())
	} else {
		for _, info := range infos {
			dataPath := filepath.Join(path, info.Name())
			if !info.IsDir() {
				if info.Name() != "iddict.csv" {
					iplocater.Error("ignore non-directory file in dataset directory: %s", dataPath)
				}
			} else if data, err := loadLocalData(dataPath); err != nil {
				iplocater.Error("ignore invalid ipip.net data directory %s: %s", dataPath, err.Error())
			} else {
				dataset.add(data)
			}
		}
	}
	return dataset, nil
}

func (dataSet *ipipNetDataSet) update() (*ipipNetData, error) {
	if response, err := http.Get(dataSet.versionURL); err != nil {
		return nil, fmt.Errorf("check ipip.net version (%s) fail: %s", dataSet.versionURL, err.Error())
	} else {
		defer response.Body.Close()
		if content, err := ioutil.ReadAll(response.Body); err != nil {
			return nil, fmt.Errorf("read ipip.net version response fail: %s", err.Error())
		} else if currentVersion := dataSet.getVersion(); currentVersion == nil || currentVersion.version != string(content) {
			version := string(content)
			path := filepath.Join(dataSet.path, version)
			if data, err := loadRemoteData(path, version, dataSet.downloadURL); err != nil {
				return nil, fmt.Errorf("load remote ipip.net data fail: %s", err.Error())
			} else {
				dataSet.add(data)
				return data, nil
			}
		} else if ok, err := currentVersion.update(dataSet.downloadURL); err != nil {
			return nil, fmt.Errorf("update current ipip.net data fail: %s", err.Error())
		} else if ok {
			return currentVersion, nil
		} else {
			return nil, nil
		}
	}
}

func (dataSet *ipipNetDataSet) Len() int {
	return len(dataSet.datas)
}

func (dataSet *ipipNetDataSet) Less(i, j int) bool {
	return dataSet.datas[i].lastModified.After(dataSet.datas[j].lastModified)
}

func (dataSet *ipipNetDataSet) Swap(i, j int) {
	dataSet.datas[i], dataSet.datas[j] = dataSet.datas[j], dataSet.datas[i]
}

func (dataSet *ipipNetDataSet) add(data *ipipNetData) {
	dataSet.datas = append(dataSet.datas, data)
	sort.Sort(dataSet)
	if dataSet.capacity > 0 {
		for len(dataSet.datas) > dataSet.capacity {
			version := dataSet.datas[len(dataSet.datas)-1]
			if err := version.remove(); err != nil {
				iplocater.Error("remove ipip.net data %s fail: %s", version.path, err.Error())
			}
			dataSet.datas = dataSet.datas[:len(dataSet.datas)-1]
		}
	}
}

func (dataSet ipipNetDataSet) getVersion() *ipipNetData {
	if dataSet.Len() == 0 {
		return nil
	} else {
		return dataSet.datas[0]
	}
}
