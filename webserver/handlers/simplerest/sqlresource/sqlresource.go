package sqlresource

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/client/sql"
	"github.com/yangchenxing/cangshan/logging"
	"github.com/yangchenxing/cangshan/webserver/handlers/simplerest"
	"strings"
)

func init() {
	application.RegisterModulePrototype("WebServerSimpleRESTSQLResource", new(Resource))
}

var (
	errGetFail    = errors.New("Get entity fail")
	errSearchFail = errors.New("Search entities fail")
	errCreateFail = errors.New("Create entity fail")
	errUpdateFail = errors.New("Update entity fail")
)

type Type interface {
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
	Fields map[string]*struct {
		Type          Type
		PrimaryKey    bool
		Creatable     bool
		Modifiable    bool
		Generater     ResourcePrimaryKeyGenerater
		AutoIncrement bool
		i             int
	}
}

func (res *Resource) Initialize() error {
	i := 0
	for _, f := range res.Fields {
		f.i = i
		i += 1
	}
	return nil
}

func (res *Resource) DecodeParams(param map[string]interface{}) error {
	for name, f := range res.Fields {
		v := param[name]
		if v == nil {
			continue
		}
		w, err := f.Type.Decode(v)
		if err != nil {
			return fmt.Errorf("Decode field %s fail: %s", name, err.Error())
		}
		param[name] = w
	}
	return nil
}

func (res *Resource) filterParam(param map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for name, _ := range res.Fields {
		if v, found := param[name]; found {
			result[name] = v
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
	valueHolders := make([]interface{}, len(res.Fields))
	conditions := make([]string, 0, 1)
	for name, f := range res.Fields {
		fieldNames[f.i] = name
		valueHolders[f.i] = f.Type.ValueHolder()
		if f.PrimaryKey {
			value, found := param[name]
			if !found {
				logging.Error("Missing %s entity primary key: %s", res.Name, name)
				return nil, errGetFail
			}
			conditions = append(conditions, name+"=?")
			sqlParams = append(sqlParams, value)
		}
	}
	statement := fmt.Sprintf("SELECT %s FROM %s WHERE %s",
		strings.Join(fieldNames, ", "), res.Name, strings.Join(conditions, " AND "))
	row := res.DB.QueryRow(statement, sqlParams...)
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
	valueHolders := make([]interface{}, len(res.Fields))
	conditions := make([]string, 0, 1)
	for name, f := range res.Fields {
		fieldNames[f.i] = name
		valueHolders[f.i] = f.Type.ValueHolder()
		if value, found := param[name]; found {
			conditions = append(conditions, name+"=?")
			if v, err := f.Type.Decode(value); err != nil {
				return nil, fmt.Errorf("Decode param %s fail: %s", name, err.Error())
			} else {
				sqlParams = append(sqlParams, v)
			}
		}
	}
	rows, err := res.DB.Query(fmt.Sprintf("SELECT %s FROM %s WHERE %s",
		strings.Join(fieldNames, ", "), res.Name, strings.Join(conditions, " AND ")), sqlParams...)
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
		result.PushBack(res.scanEntity(valueHolders))
	}
	entities := make([]map[string]interface{}, result.Len())
	for i, entity := 0, result.Front(); entity != nil; i, entity = i+1, entity.Next() {
		entities[i] = entity.Value.(map[string]interface{})
	}
	return entities, nil
}

func (res *Resource) Create(param map[string]interface{}, before, after simplerest.Trigger) (map[string]interface{}, error) {
	param = res.filterParam(param)
	if err := res.DecodeParams(param); err != nil {
		logging.Error("Decode create query param fail: %s", err.Error())
		return nil, errCreateFail
	}
	// before creation trigger
	if before != nil {
		if err := before.Handle(res.Name, nil, param); err != nil {
			logging.Error("invoke trigger before create of resource %s fail: %s", res.Name, err.Error())
			return nil, fmt.Errorf("before create trigger fail: %s", err.Error())
		}
	}

	fieldNames := make([]string, 0, len(res.Fields))
	values := make([]interface{}, 0, len(res.Fields))
	autoKeyName := ""
	keys := make(map[string]interface{})
	for name, f := range res.Fields {
		if f.PrimaryKey {
			if f.AutoIncrement {
				autoKeyName = name
			} else if f.Generater != nil {
				v, err := f.Generater.Generate()
				if err != nil {
					logging.Error("Generater resource %s primary key %s fail: %s",
						res.Name, name, err.Error())
					return nil, errCreateFail
				}
				keys[name] = v
				values = append(values, v)
				fieldNames = append(fieldNames, name)
			} else if !f.Creatable {
				logging.Error("Primary key %s of resource %s must be generatable or creatable",
					name, res.Name)
				return nil, errCreateFail
			} else if v, found := param[name]; found {
				values = append(values, v)
				fieldNames = append(fieldNames, name)
			} else {
				logging.Error("Pimary key %s of resource %s is missing", name, res.Name)
				return nil, errCreateFail
			}
		} else if v, found := param[name]; found {
			values = append(values, v)
			fieldNames = append(fieldNames, name)
		}
	}
	if autoKeyName != "" && len(keys) > 0 {
		logging.Error("Resource %s has both auto increment key and user generated key", res.Name)
		return nil, errCreateFail
	}
	if result, err := res.DB.Exec(fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)",
		res.Name, strings.Join(fieldNames, ", "),
		"?"+(strings.Repeat(" ,?", len(values)-1)))); err != nil {
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
	entity, err := res.Get(keys)
	if err != nil {
		return nil, errCreateFail
	}
	// after creation trigger
	if after != nil {
		if err := after.Handle(res.Name, nil, entity); err != nil {
			logging.Error("invoke trigger after create of resource %s fail: %s", res.Name, err.Error())
			return nil, fmt.Errorf("after create trigger fail: %s", err.Error())
		}
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
	for name, f := range res.Fields {
		if f.PrimaryKey {
			keys[name] = param[name]
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
		if err := before.Handle(res.Name, oldEntity, param); err != nil {
			logging.Error("invoke trigger before update of resource %s fail: %s", res.Name, err.Error())
			return nil, fmt.Errorf("before update trigger fail: %s", err.Error())
		}
	}

	fieldNames := make([]string, 0, len(res.Fields))
	values := make([]interface{}, 0, len(res.Fields))
	conditions := make([]string, 0, 1)
	conditionValues := make([]interface{}, 0, 1)
	for name, f := range res.Fields {
		if f.PrimaryKey {
			conditions = append(conditions, name+"=?")
			conditionValues = append(conditionValues, param[name])
		} else if v, found := param[name]; found {
			fieldNames = append(fieldNames, name+"=?")
			values = append(values, v)
		}
	}
	statement := fmt.Sprintf("UPDATE %s SET %s WHERE %s", res.Name,
		strings.Join(fieldNames, ", "), strings.Join(conditions, " AND "))
	if result, err := res.DB.Exec(statement); err != nil {
		logging.Error("Update %s entity fail: %s", res.Name, err.Error())
		return nil, errUpdateFail
	} else if count, err := result.RowsAffected(); err != nil {
		logging.Error("Update %s entity fail: cannot get count of rows affected, %s",
			res.Name, err.Error())
		return nil, errUpdateFail
	} else if count == 0 {
		logging.Warn("No %s entity updated", res.Name)
		return nil, errUpdateFail
	} else if count > 1 {
		logging.Error("More than one %s entity updated: %s", res.Name, statement)
	}
	entity, err := res.Get(keys)
	if err != nil {
		return nil, errUpdateFail
	}
	// after update trigger
	if after != nil {
		if err := after.Handle(res.Name, oldEntity, entity); err != nil {
			logging.Error("invoke trigger after update of resource %s fail: %s", res.Name, err.Error())
			return nil, fmt.Errorf("after update trigger fail: %s", err.Error())
		}
	}
	return entity, nil
}

func (res *Resource) scanEntity(values []interface{}) map[string]interface{} {
	entity := make(map[string]interface{})
	for name, f := range res.Fields {
		entity[name] = f.Type.Encode(values[f.i])
	}
	return entity
}
