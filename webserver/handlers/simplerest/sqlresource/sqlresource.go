package sqlresource

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"github.com/yangchenxing/cangshan/client/sql"
	"github.com/yangchenxing/cangshan/logging"
	"github.com/yangchenxing/cangshan/webserver/handlers/simplerest"
	"strings"
)

var (
	errGetFail    = errors.New("Get entity fail")
	errSearchFail = errors.New("Search entities fail")
	errCreateFail = errors.New("Create entity fail")
	errUpdateFail = errors.New("Update entity fail")
)

type ResourceType interface {
	ValueHolder() interface{}
	Decode(interface{}) (interface{}, error)
	Encode(interface{}) interface{}
}

type ResourcePrimaryKeyGenerater interface {
	Generate() (interface{}, error)
}

type Resource struct {
	DB     *sql.DB
	Name   string
	Fields struct {
		Name          string
		Type          ResourceType
		PrimaryKey    bool
		Creatable     bool
		Modifiable    bool
		Generater     ResourcePrimaryKeyGenerater
		AutoIncrement bool
	}
}

func (res *Resource) DecodeParams(param map[string]interface{}) error {
	for i, f := range res.Fields {
		v := param[f.Name]
		if v == nil {
			continue
		}
		w, err := f.Type.Decode(v)
		if err != nil {
			return fmt.Errorf("Decode field %s fail: %s", f.Name, err.Error())
		}
		param[f.Name] = w
	}
	return nil
}

func (res *Resource) filterParam(param map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, f := range res.Fields {
		if v, found := param[f.Name]; found {
			result[f.Name] = v
		}
	}
	return result
}

func (res *Resource) Get(param map[string]interface{}) (map[string]interface{}, error) {
	param = res.filterParam(param)
	if err := res.DecodeParams(param); err != nil {
		logging.Error("Decode get query param fail: %s", err.Error())
		return nil, errGetFail
	}
	sqlParams := make([]interface{}, 0, len(res.Fields))
	fieldNames := make([]string, len(res.Fields))
	valueHolders := make([]string, len(res.Fields))
	conditions := make([]string, 0, 1)
	for i, f := range res.Fields {
		fieldNames[i] = f.Name
		valueHolders[i] = f.Type.ValueHolder()
		if f.PrimaryKey {
			value, found := param[f.Name]
			if !found {
				logging.Error("Missing %s entity primary key: %s", res.Name, f.Name)
				return nil, errGetFail
			}
			conditions = append(conditions, f.Name+"=?")
			sqlParams = append(sqlParams, value)
		}
	}
	row := res.DB.QueryRow(fmt.Sprintf("SELECT %s FROM %s WHERE %s",
		strings.Join(fieldNames, ", "), res.Name, strings.Join(conditions, " AND ")))
	err := row.Scan(valueHolders...)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logging.Error("Scan %s entity fail: %s", res.Name, err.Error())
		return nil, errGetFail
	}
	return res.scanEntity(valueHolders), nil
}

func (res *Resource) Search(param map[string]interface{}) ([]map[string]interface{}, error) {
	param = res.filterParam(param)
	if err := res.DecodeParams(param); err != nil {
		logging.Error("Decode search query param fail: %s", err.Error())
		return nil, errSearchFail
	}
	sqlParams := make([]interface{}, 0, len(res.Fields))
	fieldNames := make([]string, len(res.Fields))
	valueHolders := make([]string, len(res.Fields))
	conditions := make([]string, 0, 1)
	for i, f := range res.Fields {
		fieldNames[i] = f.Name
		valueHolders[i] = f.Type.ValueHolder()
		if value, found := param[f.Name]; found {
			conditions = append(conditions, f.Name+"=?")
			sqlParams = append(sqlParams, f.Type.Encode(value))
		}
	}
	rows, err := res.DB.Query(fmt.Sprintf("SELECT %s FROM %s WHERE %s",
		strings.Join(fieldNames, ", "), res.Name, strings.Join(conditions, " AND ")))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logging.Error("Query %s entities fail: %s", res.Name, err.Error())
		return nil, errSearchFail
	}
	result := list.New()
	for rows.Next() {
		if err := rows.Scan(valueHolders...); err != nil {
			logging.Error("Scan %s entities fail: %s", res.Name, err.Error())
			return nil, errSearchFail
		}
		entity := res.scanEntity(valueHolders)
	}
}

func (res *Resource) Create(param map[string]interface{}, before, after simplerest.Trigger) ([]map[string]interface{}, error) {
	param = res.filterParam(param)
	if err := res.DecodeParams(param); err != nil {
		logging.Error("Decode create query param fail: %s", err.Error())
		return nil, errCreateFail
	}
	// before creation trigger
	if before != nil {
		before.Handle(res.Name, nil, param)
	}

	fieldNames := make([]string, 0, len(res.Fields))
	values := make([]interface{}, 0, len(res.Fields))
	autoKeyName := ""
	keys := make(map[string]interface{})
	for _, f := range res.Fields {
		if f.PrimaryKey {
			if f.AutoIncrement {
				autoKeyName = f.Name
			} else if f.Generater != nil {
				v, err := f.Generater.Generate()
				if err != nil {
					logging.Error("Generater resource %s primary key %s fail: %s",
						res.Name, f.Name, err.Error())
					return nil, errCreateFail
				}
				keys[f.Name] = v
				values = append(values, v)
				fieldNames = append(fieldNames, f.Name)
			} else if !f.Creatable {
				logging.Error("Primary key %s of resource %s must be generatable or creatable",
					f.Name, res.Name)
				return nil, errCreateFail
			} else if v, found := param[f.Name]; found {
				values = append(values, v)
				fieldNames = append(fieldNames, f.Name)
			} else {
				logging.Error("Pimary key %s of resource %s is missing".f.Name, res.Name)
				return nil, errCreateFail
			}
		} else if v, found := param[f.Name]; found {
			values = append(values, v)
			fieldNames = append(fieldNames, f.Name)
		}
	}
	if autoKeyName != "" && len(keys) > 0 {
		logging.Error("Resource %s has both auto increment key and user generated key", res.Name)
		return errCreateFail
	}
	if result, err := res.DB.Exec(fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)",
		res.Name, strings.Join(fieldNames, ", "),
		"?"+strings.Join(strings.Repeat(" ,?", len(values)-1)))); err != nil {
		logging.Error("Insert %s entity fail: %s", res.Name, err.Error())
		return nil, errCreateFail
	} else if autoKeyName != "" {
		key, err := result.LastInsertId()
		if err != nil {
			logging.Error("Get %s entity auto increment key fail: %s", res.Name, err.Error())
			return nil, errCreateFail
		}
		keys[autoKeyName] = key
	}
	if entity, err := res.Get(keys); err == nil {
		return entity, nil
	}
	// after creation trigger
	if after != nil {
		after.Handle(res.Name, nil, entity)
	}
	return nil, errCreateFail
}

func (res *Resource) Update(param map[string]interface{}, before, after simplerest.Trigger) (map[string]interface{}, error) {
	param = res.filterParam(param)
	if err := res.DecodeParams(param); err != nil {
		logging.Error("Decode update query param fail: %s", err.Error())
		return nil, errUpdateFail
	}
	// prepare primary key
	keys := make(map[string]interface{})
	for _, f := range res.Fields {
		if f.PrimaryKey {
			keys[f.Name] = param[f.Name]
		}
	}
	var oldEntity map[string]interface{}
	// prepare old entity for triggers
	if before != nil || after != nil {
		var err error
		if oldEntity, err = res.Get(param); err != nil {
			logging.Error("Get old %s entity %v fail: %s", res.Name, keys, err.Error())
			return nil, errCreateFail
		}
	}
	// before update trigger
	if before != nil {
		before(oldEntity, param)
	}

	fieldNames := make([]string, 0, len(res.Fields))
	values := make([]interface{}, 0, len(res.Fields))
	conditions := make([]string, 0, 1)
	conditionValues := make([]interface{}, 0, 1)
	for _, f := range res.Fields {
		if f.PrimaryKey {
			conditions = append(conditions, f.Name+"=?")
			conditionValues = append(keys, param[f.Name])
		} else if v, found := param[f.Name]; found {
			fieldNames = append(fieldNames, f.Name+"=?")
			values = append(values)
		}
	}
	if result, err := res.DB.Exec(fmt.Sprintf("UPDATE %s SET %s WHERE %s", res.Name,
		strings.Join(fieldNames, ", "), strings.Join(conditions, " AND "))); err != nil {
		logging.Error("Update %s entity fail: %s", res.Name, err.Error())
		return nil, errUpdateFail
	}
	entity, err := res.Get(keys)
	if err != nil {
		return nil, errUpdateFail
	}
	// after update trigger
	if after != nil {
		after.Handle(res.Name, oldEntity, entity)
	}
	return entity, nil
}

func (res *Resource) scanEntity(values []interface{}) map[string]interface{} {
	entity := make(map[string]interface{})
	for i, f := range res.Fields {
		entity[f.Name] = f.Type.Encode(values[i])
	}
	return entity
}
