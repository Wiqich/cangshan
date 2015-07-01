package webserver

import (
	"encoding/json"
	"github.com/yangchenxing/cangshan/logging"
)

func WriteStandardJSONResult(request *Request, success bool, key string, value interface{}) {
	result := map[string]interface{}{
		"success": success,
	}
	if key != "" {
		result[key] = value
	}
	content, err := json.Marshal(result)
	if err != nil {
		logging.Error("Marshal standard json success entity fail: %s", err.Error())
		request.Write(500, nil, "")
	} else {
		request.Write(200, content, "application/json")
	}
}
