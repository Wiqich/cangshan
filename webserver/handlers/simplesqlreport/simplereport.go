package simplereport

import (
	"fmt"
	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/client/sql"
	"github.com/yangchenxing/cangshan/webserver"
	"github.com/yangchenxing/cangshan/webserver/handlers/simplerest/sqlresource"
)

func init() {
	application.RegisterModulePrototype("WebServerSimpleReport", new(SimpleReport))
}

type SimpleReport struct {
	DB     *sql.DB
	SQL    string
	Params []struct {
		Name string
		Type sqlresource.Type
	}
	Fields map[string]struct {
		Name string
		Type sqlresource.Type
	}
	GroupKey string
}

func (report *SimpleReport) Handle(request *webserver.Request) {
	params := make([]interface{}, len(report.Params))
	for i, param := range report.Params {
		if p := request.Params[param.Name]; p == nil {
			webserver.WriteStandardJSONResult(request, false, "message", "缺少"+param.Name)
			return
		} else if v, err := param.Type.Decode(p); err != nil {
			webserver.WriteStandardJSONResult(request, false, "message", fmt.Sprintf("参数%s错误", param.Name))
			return
		} else {
			params[i] = v
		}
	}
	rows, err := report.DB.Query(report.SQL, params)
	if err != nil {
		request.Error("查询报表出错: %s", err.Error())
		webserver.WriteStandardJSONResult(request, false, "message", "服务器内部错误")
		return
	}
	row := make([]interface{}, len(report.Fields))
	for i, field := range report.Fields {
		row[i] = field.Type.ValueHolder()
	}
	records := make([]map[string]interface{}, 0, 32)
	for rows.Next() {
		if err := rows.Scan(row...); err != nil {
			request.Error("查询报表出错: %s", err.Error())
			webserver.WriteStandardJSONResult(request, false, "message", "服务器内部错误")
			return
		}
		record := make(map[string]interface{})
		for i, field := range report.Fields {
			record[field.Name] = field.Type.Encode(row[i])
		}
		records = append(records, record)
	}
	if report.GroupKey != "" {
		groups := make(map[string][]map[string]interface{})
		for _, record := range records {
			key := record[report.GroupKey]
			group, found := groups[key]
			if !found {
				group = make([]map[string]interface{}, 0, 32)
			}
			groups[key] = append(group, record)
		}
		webserver.WriteStandardJSONResult(request, true, "entities", groups)
	} else {
		webserver.WriteStandardJSONResult(request, true, "entities", records)
	}
}
