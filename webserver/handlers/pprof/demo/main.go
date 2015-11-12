package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/application/config/toml"
	"github.com/yangchenxing/cangshan/logging"
	_ "github.com/yangchenxing/cangshan/webserver/handlers/pprof"
)

var (
	config = flag.String("config", "conf/config.toml", "配置文件路径")
)

func exit(code int) {
	time.Sleep(time.Millisecond)
	logging.Flush()
	os.Exit(code)
}

func main() {
	flag.Parse()
	app, err := tomlapp.NewApplication(*config)
	if err == application.ErrDeadlock {
		logging.Error("创建应用失败: 模块依赖死锁")
		for _, seq := range app.DumpWatingSequences() {
			logging.Warn("模块依赖: %s", strings.Join(seq, " -> "))
		}
		exit(1)
	}
	if err != nil {
		logging.Error("创建应用失败: %s", err.Error())
		exit(1)
	}
	if err = app.Run(); err != nil {
		logging.Error("运行应用失败: %s", err.Error())
		exit(2)
	}
}
