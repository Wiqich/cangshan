package webserver

import (
	"encoding/json"
	"github.com/yangchenxing/cangshan/logging"
)

func WriteStandardJSONResult(request *Request, success bool, params ...interface{}) {
	result := map[string]interface{}{
		"success": success,
	}
	for i := 0; i+1 < len(params); i += 2 {
		if key, ok := params[i].(string); ok {
			result[key] = params[i+1]
		}
	}
	content, err := json.Marshal(result)
	if err != nil {
		logging.Error("Marshal standard json success entity fail: %s", err.Error())
		request.Write(500, nil, "")
	} else {
		request.Write(200, content, "application/json")
	}
}
