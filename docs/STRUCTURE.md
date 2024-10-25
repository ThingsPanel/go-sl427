# 项目结构说明

## 目录结构

```go
go-sl427/
├── cmd/                    # 命令行工具
│   └── examples/          # 示例程序
├── docs/                  # 文档
│   ├── DEVELOPMENT.md    # 开发计划文档
│   ├── STRUCTURE.md      # 项目结构说明
│   └── PROTOCOL.md       # 协议细节文档
├── pkg/
│   └── sl427/            # 核心包
│       ├── protocol/     # 协议实现
│       │   ├── packet/   # 报文处理
│       │   ├── types/    # 数据类型
│       │   └── codec/    # 编解码
│       ├── station/      # 站点相关
│       └── transport/    # 传输层
├── test/                 # 测试用例
└── README.md             # 项目说明
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
   - 维持心跳连接

5. **transport包**
   - 处理网络传输
   - 管理连接状态
   - 提供消息处理
