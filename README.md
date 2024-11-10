# go-sl427

[![License](https://img.shields.io/github/license/ThingsPanel/go-sl427)](LICENSE)
[![Release](https://img.shields.io/github/release/ThingsPanel/go-sl427.svg)](https://github.com/ThingsPanel/go-sl427/releases)
[![GitHub Issues](https://img.shields.io/github/issues/ThingsPanel/go-sl427.svg)](https://github.com/ThingsPanel/go-sl427/issues)
[![Go Report Card](https://goreportcard.com/badge/github.com/ThingsPanel/go-sl427)](https://goreportcard.com/report/github.com/ThingsPanel/go-sl427)
[![Go Reference](https://pkg.go.dev/badge/github.com/ThingsPanel/go-sl427.svg)](https://pkg.go.dev/github.com/ThingsPanel/go-sl427)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ThingsPanel/go-sl427)](https://go.dev/doc/devel/release)

go-sl427是一个用Go语言实现的SL427-2021水资源监测数据传输规约库。该库提供了协议实现，支持监测站和数据中心服务器的开发。

## 特性

- 实现SL427-2021的S1链路协议规范
- 支持监测站和服务器端开发
- 提供灵活的配置选项
- 内置监控指标收集
- 支持自定义日志接口
- 线程安全设计
- 详细的错误处理

## 完整性说明

目前完成了协议框架的实现
具体功能待完善

- [x] S1 链路协议
- [ ] S2 链路协议
- [ ] S3 链路协议

## 安装

```bash
go get github.com/ThingsPanel/go-sl427
```

## 快速开始

### 监测站示例

### 服务器示例

## 核心组件

- **station**: 监测站实现
- **transport**: 网络传输层
- **codec**: 数据编解码
- **packet**: 数据包定义
- **types**: 基础类型定义
- **metrics**: 监控指标收集

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

在 `cmd/examples` 目录下提供了完整的示例程序


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
