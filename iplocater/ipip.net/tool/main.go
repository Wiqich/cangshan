package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/yangchenxing/cangshan/iplocater/ipip.net"
	"github.com/yangchenxing/cangshan/logging"
)

var (
	path           = flag.String("path", "data/ipip.net", "data path")
	versionURL     = flag.String("version", "", "ipip.net version url")
	downloadURL    = flag.String("download", "", "ipip.net download url")
	updateInterval = flag.Duration("update", time.Minute, "ipip.net update interval")
)

func exit(code int) {
	logging.Flush()
	os.Exit(code)
}

func main() {
	logging.CreateDefaultLogging()
	flag.Parse()
	if *versionURL == "" {
		logging.Error("\"version\" is required argument")
		exit(1)
	} else if *downloadURL == "" {
		logging.Error("\"download\" is required argument")
		exit(1)
	}
	locater := &ipipnet.IPIPNet{
		VersionURL:     *versionURL,
		DownloadURL:    *downloadURL,
		Path:           *path,
		UpdateInterval: *updateInterval,
	}
	if err := locater.Initialize(); err != nil {
		fmt.Println("初始化ipip.net定位器出错:", err.Error())
		os.Exit(1)
	}
	logging.Debug("初始化成功")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		if ip := net.ParseIP(scanner.Text()); ip == nil {
			fmt.Println("错误的IP:", scanner.Text())
			continue
		} else if location, err := locater.Locate(ip); err != nil {
			fmt.Println("定位出错:", err.Error())
			continue
		} else if text, err := json.Marshal(location); err != nil {
			fmt.Println("序列化定位结果出错:", err.Error())
			continue
		} else {
			fmt.Println("定位结果:", ip, "@", string(text))
		}
	}
}
