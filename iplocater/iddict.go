package iplocater

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type regionRecord struct {
	*Region
	subregions map[string]regionRecord
}

type IDDict struct {
	Regions  map[string]regionRecord
	ISPs     map[string]*ISP
	unknowns map[string]bool
}

func LoadIDDict(path string) (*IDDict, error) {
	dict := &IDDict{
		Regions:  make(map[string]regionRecord),
		ISPs:     make(map[string]*ISP),
		unknowns: make(map[string]bool),
	}
	if file, err := os.Open(path); err != nil {
		return nil, err
	} else {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			record := strings.Split(scanner.Text(), ",")
			// fmt.Println("iddict record:", record)
			switch record[0] {
			case "region":
				id, err := strconv.ParseUint(record[len(record)-1], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid region line: %s", strings.Join(record, ","))
				}
				if len(record) >= 2 && len(record) <= 5 {
					if err = dict.setRegionID(id, record[1:len(record)-1]); err != nil {
						return nil, fmt.Errorf("invalid region line: %s", err.Error())
					}
				} else {
					return nil, fmt.Errorf("invalid region line: %s", strings.Join(record, ","))
				}
			case "isp":
				id, err := strconv.ParseUint(record[len(record)-1], 10, 32)
				if err != nil {
					return nil, fmt.Errorf("invalid isp line: %s", strings.Join(record, ","))
				}
				switch len(record) - 1 {
				case 2:
					if _, found := dict.ISPs[record[1]]; found {
						return nil, fmt.Errorf("duplicated ISP: %s", record[1])
					} else {
						dict.ISPs[record[1]] = &ISP{
							ID:   id,
							Name: record[1],
						}
					}
				default:
					return nil, fmt.Errorf("invalid isp line: %s", strings.Join(record, ","))
				}
			default:
				return nil, fmt.Errorf("invalid isp line: %s", strings.Join(record, ","))
			}
		}
	}
	return dict, nil
}

func (dict *IDDict) setRegionID(id uint64, names []string) error {
	if len(names) > 3 {
		return fmt.Errorf("over max region depth: %d>3", len(names))
	}
	regions := dict.Regions
	depth := len(names) - 1
	for i := 0; i < depth; i++ {
		if _, found := regions[names[i]]; !found {
			return fmt.Errorf("unknown region: %s", strings.Join(names[:i+1], ","))
		} else {
			regions = regions[names[i]].subregions
		}
	}
	if _, found := regions[names[depth]]; found {
		return fmt.Errorf("duplicated regions: %s", strings.Join(names, ","))
	}
	r := regionRecord{
		Region: &Region{
			ID:   id,
			Name: names[depth],
		},
	}
	if depth < 2 {
		r.subregions = make(map[string]regionRecord)
	}
	regions[names[depth]] = r
	return nil
}

func (dict *IDDict) GetLocation(regionNames []string, ispNames []string) *Location {
	location := new(Location)
	regions := dict.Regions
	for i, name := range regionNames {
		if i > 2 || name == "" || (i == 1 && name == regionNames[0]) {
			break
		}
		if len(regions) == 0 {
			break
		}
		region, found := regions[name]
		if !found {
			if key := strings.Join(regionNames[:i+1], "/"); !dict.unknowns[key] {
				dict.unknowns[key] = true
				if i > 0 {
					Debug("unknown level %d region: %s", i, key)
				}
			}
			break
		}
		switch i {
		case 0:
			location.Country = region.Region
		case 1:
			location.Province = region.Region
		case 2:
			location.City = region.Region
		}
		regions = region.subregions
	}
	location.ISPs = make([]*ISP, 0, len(ispNames))
	for _, name := range ispNames {
		if name == "" {
			continue
		}
		isp := dict.ISPs[name]
		if isp != nil {
			location.ISPs = append(location.ISPs, isp)
		} else if !dict.unknowns[name] {
			dict.unknowns[name] = true
			// Debug("unknown isp: %s", name)
		}
	}
	return location
}
