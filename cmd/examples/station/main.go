// cmd/examples/station/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThingsPanel/go-sl427/pkg/sl427/station"
)

func main() {
	// 解析命令行参数
	var (
		serverAddr     string
		stationAddrHex string // 使用16进制字符串接收站点地址
		interval       time.Duration
	)

	flag.StringVar(&serverAddr, "server", "localhost:8080", "服务器地址")
	flag.StringVar(&stationAddrHex, "station", "12345678", "站点地址(16进制)")
	flag.DurationVar(&interval, "interval", 10*time.Second, "数据上报间隔")
	flag.Parse()

	// 解析16进制站点地址
	var stationAddr uint32
	_, err := fmt.Sscanf(stationAddrHex, "%x", &stationAddr)
	if err != nil {
		log.Fatalf("无效的站点地址: %v", err)
	}

	// 创建站点配置
	config := station.Config{
		Address:  stationAddr,
		Server:   serverAddr,
		Interval: interval,
	}

	// 创建站点实例
	st := station.NewStation(config)

	// 设置日志
	st.SetLogger(log.Default())

	// 启动站点
	if err := st.Start(config); err != nil {
		log.Printf("站点启动失败: %v", err)
		os.Exit(1)
	}

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("站点已启动 [地址: %X] [服务器: %s] [上报间隔: %v]",
		stationAddr, serverAddr, interval)
	log.Printf("按 Ctrl+C 停止...")

	<-sigChan

	// 优雅退出
	st.Stop()
	log.Printf("站点已停止")
}
