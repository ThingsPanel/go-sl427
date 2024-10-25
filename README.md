# go-sl427

[![Go Reference](https://pkg.go.dev/badge/github.com/ThingsPanel/go-sl427.svg)](https://pkg.go.dev/github.com/ThingsPanel/go-sl427)
[![Go Report Card](https://goreportcard.com/badge/github.com/ThingsPanel/go-sl427)](https://goreportcard.com/report/github.com/ThingsPanel/go-sl427)

go-sl427是一个用Go语言实现的SL427-2021水资源监测数据传输规约库。该库提供了完整的协议实现，支持监测站和数据中心服务器的开发。

## 特性

- 完整实现SL427-2021协议规范
- 支持监测站和服务器端开发
- 提供灵活的配置选项
- 内置监控指标收集
- 支持自定义日志接口
- 线程安全设计
- 详细的错误处理

## 安装

```bash
go get github.com/ThingsPanel/go-sl427
```

## 快速开始

### 监测站示例

```go
package main

import (
    "log"
    "time"
    
    "github.com/ThingsPanel/go-sl427/pkg/sl427/station"
)

func main() {
    // 创建监测站实例
    config := station.Config{
        Address:  0x01,           // 站点地址
        Server:   "localhost:8080", // 服务器地址
        Interval: time.Second * 30, // 数据上报间隔
    }
    
    s := station.NewStation(config)
    
    // 启动监测站
    if err := s.Start(config); err != nil {
        log.Fatal(err)
    }
    defer s.Stop()
    
    // 保持运行
    select {}
}
```

### 服务器示例

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/ThingsPanel/go-sl427/pkg/sl427/transport"
)

func main() {
    // 服务器配置
    config := transport.Config{
        ListenAddr:    ":8080",
        ReadTimeout:   30,
        WriteTimeout:  30,
        MaxConns:      1000,
        MaxPacketSize: 1024,
    }
    
    // 创建服务器
    server := transport.NewServer(config)
    
    // 创建context用于优雅关闭
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // 启动服务器
    if err := server.Start(ctx); err != nil {
        log.Fatal(err)
    }
    
    // 等待中断信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    
    // 优雅关闭
    server.Stop()
}
```

## 核心组件

- **station**: 监测站实现
- **transport**: 网络传输层
- **protocol**: 协议解析与封装
- **codec**: 数据编解码
- **packet**: 数据包定义
- **types**: 基础类型定义
- **metrics**: 监控指标收集

## 数据项定义

库提供了内置的数据项注册表，支持自定义数据项定义：

```go
// 注册数据项
types.DefaultRegistry.RegisterBatch([]types.DataItemDef{
    {
        ID:          1001,
        Name:        "水位",
        Type:        types.TypeInt32,
        Unit:        "m",
        Scale:       -3,
        Description: "站点水位",
    },
    // ... 更多数据项定义
})
```

## 监控指标

内置的监控指标包括：

- 接收的数据包数量
- 发送的数据包数量
- 丢弃的数据包数量
- 最后接收时间
- 最后发送时间
- 处理延迟

## 错误处理

库提供了统一的错误处理机制：

```go
if sl427.IsErrorCode(err, sl427.ErrCodeInvalidData) {
    // 处理无效数据错误
}
```

## 配置选项

both站点和服务器支持多种配置选项：

```go
// 站点配置
station.Config{
    Address:  0x01,
    Server:   "localhost:8080",
    Interval: time.Second * 30,
}

// 服务器配置
transport.Config{
    ListenAddr:    ":8080",
    ReadTimeout:   30,
    WriteTimeout:  30,
    MaxConns:      1000,
    MaxPacketSize: 1024,
}
```

## 日志接口

支持自定义日志实现：

```go
type CustomLogger struct {
    // 自定义日志实现
}

func (l *CustomLogger) Printf(format string, v ...interface{}) {
    // 实现日志记录
}

// 设置日志接口
types.SetLogger(&CustomLogger{})
```

## 示例程序

在 `cmd/examples` 目录下提供了完整的示例程序：

- `basic`: 基础使用示例
- `server`: 完整的服务器示例
- `station`: 监测站示例

## 文档

详细的文档和API参考请访问：[pkg.go.dev](https://pkg.go.dev/github.com/ThingsPanel/go-sl427)

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开Pull Request

## 许可证

采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 支持

如有问题或建议，请提交 [Issue](https://github.com/ThingsPanel/go-sl427/issues)。

## 致谢

感谢所有贡献者对项目的支持。