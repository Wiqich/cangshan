package iplocater

import (
	"encoding/csv"
	"os"
)

type country struct {
	*Location
	provinces map[string]province
}

type province struct {
	*Location
	cities map[string]city
}

type city struct {
	*Location
}

type isp struct {
	id   uint32
	name string
}

type IDDict struct {
	countries map[string]country
	isps      map[string]isp
}

func LoadIDDict(path string) (*IDDict, error) {
	dict := &IDDict{
		countries: make(map[string]country),
		isps:      make(map[string]isp),
	}
	if file, err := os.Open(path); err != nil {
		return nil, err
	} else {
		reader := csv.NewReader(r)
		for record, err := reader.Read(); record != nil && err == nil; record, err = reader.Read() {
			//
		}
	}
}
