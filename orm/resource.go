package orm

import (
	"bytes"
	"container/list"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/yangchenxing/cangshan/client/sql"
)

type DB interface {
	Exec(string ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
}

type FieldValidator interface {
	Validate(interface{}) bool
}

type EntityValidator interface {
	Validate(map[string]interface{}) bool
}

type Field struct {
	Name            string
	DBName          string
	Type            string
	PrimaryKey      bool
	Nullable        bool
	OmitNull        bool
	Modifiable      bool
	Hidden          bool
	FieldValidators []FieldValidator
}

func (field Field) NewValueHolder() interface{} {
	switch field.Type {
	case "int":
		if field.Nullable {
			return new(sql.NullInt64)
		}
		return new(int64)
	case "string":
		if field.Nullable {
			return new(sql.NullString)
		}
		return new(string)
	case "float":
		if field.Nullable {
			return new(sql.NullFloat64)
		}
		return new(float64)
	case "bool":
		if field.Nullable {
			return new(sql.NullBool)
		}
		return new(bool)
	case "time":
		if field.Nullable {
			return new(sql.NullTime)
		}
		return new(time.Time)
	}
	return nil
}

type Entity struct {
	values   map[string]interface{}
	modified map[string]bool
	resource *Resource
	db       DB
}

func (entity Entity) Get(name string) interface{} {
	return entity.values[name]
}

func (entity *Entity) Set(name string, value interface{}) {
	if entity.modified == nil {
		entity.modified = make(map[string]bool)
	}
	if entity.resource.getField(name) != nil {
		entity.values[name] = value
		entity.modified[name] = true
	}
}

func (entity *Entity) Sync() error {

}

type Resource struct {
	DB               *sql.DB
	TableName        string
	Fields           []*Field
	EntityValidators []EntityValidator
	fields           map[string]*Field
}

func (resource Resource) getField(name string) *Field {
	if resource.fields == nil {
		resource.fields = make(map[string]*Field)
		for _, field := range resource.Fields {
			resource.fields[field.Name] = field
		}
	}
	return resource.fields[name]
}

func (resource *Resource) NewQuery() *Query {
	return &Query{
		Condition: newCondition(),
		resource:  resource,
	}
}

type Query struct {
	*Condition
	resource    *Resource
	order       *list.List
	limitOffset int
	limitCount  int
}

func (query *Query) Order(name string, order string) *Query {
	if query.order == nil {
		query.order = list.New()
	}
	if order = strings.ToUpper(order); order == "ASC" || order == "DESC" {
		query.order.PushBack([2]string{name, order})
	}
	return query
}

func (query *Query) Limit(offset int, count int) *Query {
	query.limitOffset = offset
	query.limitCount = count
	return query
}

func (query *Query) Fetch() ([]*Entity, error) {
	var buffer bytes.Buffer
	argList := list.New()
	fields, valueHolders := query.buildSelectClause(&buffer, argList)
	query.buildFromClause(&buffer, argList)
	query.buildWhereClause(&buffer, argList)
	query.buildOrderClause(&buffer, argList)
	query.buildLimitClause(&buffer, argList)
	args := make([]interface{}, argList.Len())
	for i, elem := 0, argList.Front(); elem != nil; i, elem = i+1, elem.Next() {
		args[i] = elem.Value
	}
	entityList := list.New()
	if rows, err := query.resource.DB.Query(buffer.String(), args...); err != nil {
		return nil, err
	} else {
		for rows.Next() {
			if err := rows.Scan(valueHolders); err != nil {
				return nil, err
			}
			entity := &Entity{
				values:   make(map[string]interface{}),
				resource: query.resource,
			}
			for i := range fields {
				value := field.get(valueHolders[i])
				if value != nil || field.OmitNull {
					entity.values[field.Name] = value
				}
			}
		}
	}
	entities := make([]*Entity, entityList.Len())
	for i, elem := 0, entityList.Front(); elem != nil; i, elem = i+1, elem.Next() {
		entities[i] = elem.Value.(*Entity)
	}
	return entities, nil
}

func (query *Query) FetchOne() (*Entity, error) {
	var buffer bytes.Buffer
	fields, valueHolders := query.buildSelectClause(&buffer)
	query.buildFromClause(&buffer)
	args := query.buildWhereClause(&buffer)
	if row, err := query.resource.DB.QueryRow(buffer.String(), args...); err != nil {
		return nil, err
	} else if err := row.Scan(valueHolders); err != nil {
		return nil, err
	}
	entity := &Entity{
		values:   make(map[string]interface{}),
		resource: query.resource,
	}
	for i := range fields {
		value := field.get(valueHolders[i])
		if value != nil || field.OmitNull {
			entity.values[field.Name] = value
		}
	}
	return entity, nil
}

func (query *Query) Create(map[string]interface{}) (*Entity, error) {

}

func (query *Query) buildSelectClause(writer io.Writer) ([]*Field, []interface{}) {
	fieldList := list.New()
	for _, field := range query.resource.Fields {
		if !field.Hidden {
			fieldList.PushBack(field)
		}
	}
	fields := make([]*Field, fieldList.Len())
	args := make([]interface{}, fieldList.Len())
	for i, field := 0, fieldList.Front(); field != nil; i, field = i+1, field.Next() {
		fields[i] = field.Value().(*Field)
		args[i] = field.Valule().(*Field).NewValueHolder()
	}
	for i, field := range fields {
		name := field.DBName
		if name == "" {
			name == field.Name
		}
		if i == 0 {
			fmt.Fprintf(writer, "SELECT %s", name)
		} else {
			fmt.Fprintf(writer, ", %s", name)
		}
	}
	return fields, args
}

func (query *Query) buildFromClause(writer io.Writer) {
	fmt.Fprintf(writer, " FROM %s", query.resource.TableName)
}

func (query *Query) buildWhereClause(writer io.Writer) []interface{} {
	if condition, args := query.condition.build(); condition != "" {
		fmt.Fprintf(writer, " WHERE %s", condition)
	}
	return args
}
