package experiment

import (
	"encoding/json"
	"fmt"
	"github.com/yangchenxing/cangshan/logging"
)

func LoadJson(data []byte) (*Branch, error) {
	root := new(Branch)
	if err := json.Unmarshal(data, root); err != nil {
		return nil, err
	}
	return root, root.Initialize()
}

type FeatureMark struct {
	Name      string
	Value     interface{}
	Operation string
}

type Branch struct {
	Name     string
	Router   map[string]interface{}
	Marks    []FeatureMark
	Children []Branch
	router   Router
}

func (branch *Branch) Initialize() error {
	var err error
	if branch.router, err = createRouter(branch.Router); err != nil {
		return err
	}
	for i, _ := range branch.Children {
		if err = branch.Children[i].Initialize(); err != nil {
			return err
		}
	}
	return nil
}

func (branch *Branch) Decide(features Features) error {
	logging.Debug("进入实验节点: %s", branch.Name)
	if branch.Marks != nil {
		for _, mark := range branch.Marks {
			features.UpdateFeature(mark.Name, mark.Value, mark.Operation)
		}
	}
	if branch.router != nil {
		if choice, err := branch.router.SelectBranch(features); err != nil {
			return fmt.Errorf("分支决策出错: name=%s, error=\"%s\"", branch.Name, err.Error())
		} else if choice >= len(branch.Children) {
			return fmt.Errorf("决策分支超边界: name=%s, choice=%d, children=%d",
				branch.Name, choice, len(branch.Children))
		} else if choice >= 0 {
			return branch.Children[choice].Decide(features)
		}
	} else {
		for i, _ := range branch.Children {
			if err := branch.Children[i].Decide(features); err != nil {
				return err
			}
		}
	}
	return nil
}
