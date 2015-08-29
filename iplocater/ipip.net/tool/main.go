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
	path = flag.String("path", "data/ipip.net", "IP库数据路径")
)

func main() {
	logging.CreateDefaultLogging()
	flag.Parse()
	locater := &ipipnet.IPIPNet{
		VersionURL:     "http://user.ipip.net/download.php?a=version&token=f41859e64780afe1772df0e73b0f7ce4de8a514c",
		DownloadURL:    "http://user.ipip.net/download.php?token=f41859e64780afe1772df0e73b0f7ce4de8a514c",
		Path:           *path,
		UpdateInterval: time.Hour,
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
