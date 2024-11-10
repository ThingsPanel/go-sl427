# 项目结构说明

## 目录结构

```go
go-sl427/
├─.github
├─cmd
│  ├─examples
│  │  └─basic
│  └─merger
├─docs
│  └─design
│      └─dataflow
└─pkg
    └─sl427
        ├─codec
        ├─metrics
        ├─packet
        ├─protocol
        ├─station
        ├─transport
        └─types
```

## 结构说明

### 核心包结构

1. **types包**
   - 实现基础数据类型
   - 定义错误类型
   - 提供类型转换

2. **packet包**
   - 实现报文结构
   - 提供报文构建方法
   - 处理报文校验

3. **codec包**
   - 提供编解码能力
   - 处理数据转换
   - 实现CRC校验

4. **station包**
   - 实现测站功能
   - 处理数据采集

5. **transport包**
   - 处理网络传输
   - 管理连接状态
   - 提供消息处理
