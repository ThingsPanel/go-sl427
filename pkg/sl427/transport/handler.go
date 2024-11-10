// pkg/sl427/transport/handler.go
package transport

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"

	"github.com/ThingsPanel/go-sl427/pkg/sl427"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/codec"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/packet"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
)

// Handler 定义数据包处理器接口
type Handler interface {
	// Handle 处理连接
	Handle() error

	// SetLogger 设置日志接口
	SetLogger(logger types.Logger)

	// Close 关闭处理器
	Close() error

	// RemoteAddr 获取远程地址
	RemoteAddr() net.Addr
}

// PacketHandler 包处理器接口
type PacketHandler interface {
	// HandlePacket 处理单个数据包
	HandlePacket(*packet.Packet) error
}

// HandlerConfig 处理器配置
type HandlerConfig struct {
	MaxPacketSize int          // 最大包大小
	ReadTimeout   int          // 读超时(秒)
	WriteTimeout  int          // 写超时(秒)
	Logger        types.Logger // 日志接口
}

// Option 处理器配置选项
type Option func(*HandlerConfig)

// WithMaxPacketSize 设置最大包大小
func WithMaxPacketSize(size int) Option {
	return func(c *HandlerConfig) {
		c.MaxPacketSize = size
	}
}

// WithLogger 设置日志接口
func WithLogger(logger types.Logger) Option {
	return func(c *HandlerConfig) {
		c.Logger = logger
	}
}

// WithTimeout 设置超时时间
func WithTimeout(readTimeout, writeTimeout int) Option {
	return func(c *HandlerConfig) {
		c.ReadTimeout = readTimeout
		c.WriteTimeout = writeTimeout
	}
}

// DefaultConfig 默认配置
var DefaultConfig = HandlerConfig{
	MaxPacketSize: 1024,
	ReadTimeout:   30,
	WriteTimeout:  30,
	Logger:        types.DefaultLogger,
}

// handlerImpl 处理器实现
type handlerImpl struct {
	conn          net.Conn
	config        HandlerConfig
	codec         *codec.PacketCodec
	reader        *bufio.Reader
	logger        types.Logger
	packetHandler PacketHandler
}

// NewHandler 创建新的连接处理器
func NewHandler(conn net.Conn, handler PacketHandler, opts ...Option) Handler {
	config := DefaultConfig

	// 应用配置选项
	for _, opt := range opts {
		opt(&config)
	}

	return &handlerImpl{
		conn:          conn,
		config:        config,
		codec:         codec.NewPacketCodec(),
		reader:        bufio.NewReader(conn),
		logger:        config.Logger,
		packetHandler: handler,
	}
}

// Handle 实现Handler接口：处理连接
func (h *handlerImpl) Handle() error {
	defer h.Close()

	h.logger.Printf("新连接建立: %s", h.conn.RemoteAddr())

	for {
		// 读取并处理数据包
		p, err := h.readPacket()
		if err != nil {
			if err != io.EOF {
				h.logger.Printf("读取数据失败 [%s]: %v", h.conn.RemoteAddr(), err)
				if sl427.IsErrorCode(err, sl427.ErrCodeInvalidData) {
					continue // 尝试重新同步
				}
				return err
			}
			return nil // 连接正常关闭
		}

		// 处理数据包
		if err := h.packetHandler.HandlePacket(p); err != nil {
			h.logger.Printf("处理数据包失败 [%s]: %v", h.conn.RemoteAddr(), err)
			continue
		}
	}
}

// pkg/sl427/server/handler.go
func (h *handlerImpl) readPacket() (*packet.Packet, error) {
	var buf bytes.Buffer

	// 1. 查找起始标识
	startByte, err := h.reader.ReadByte()
	if err != nil {
		return nil, sl427.WrapError(sl427.ErrCodeConnectionFailed, "读取起始字节失败", err)
	}
	buf.WriteByte(startByte)

	// 确保是起始字节
	if startByte != types.StartFlag {
		return nil, sl427.WrapError(sl427.ErrCodeInvalidData, "无效的起始标识", nil)
	}

	// 2. 读取长度字节
	length, err := h.reader.ReadByte()
	if err != nil {
		return nil, sl427.WrapError(sl427.ErrCodeConnectionFailed, "读取长度字节失败", err)
	}
	buf.WriteByte(length)

	// 3. 读取第二个起始标识
	startByte2, err := h.reader.ReadByte()
	if err != nil {
		return nil, sl427.WrapError(sl427.ErrCodeConnectionFailed, "读取第二个起始标识失败", err)
	}
	buf.WriteByte(startByte2)

	if startByte2 != types.StartFlag {
		return nil, sl427.WrapError(sl427.ErrCodeInvalidData, "无效的第二个起始标识", nil)
	}

	// 4. 计算需要读取的剩余字节数
	// 总长度 = 用户数据区长度 + 帧头(3) + CS(1) + 结束符(1)
	remainingBytes := int(length) + 2 // +2是CS和结束符

	// 5. 读取剩余数据
	data := make([]byte, remainingBytes)
	n, err := io.ReadFull(h.reader, data)
	if err != nil {
		return nil, sl427.WrapError(sl427.ErrCodeConnectionFailed, "读取剩余数据失败", err)
	}
	if n != remainingBytes {
		return nil, sl427.WrapError(sl427.ErrCodeInvalidLength,
			fmt.Sprintf("数据长度不匹配,期望:%d,实际:%d", remainingBytes, n), nil)
	}
	buf.Write(data)

	// 6. 使用codec解码完整的帧
	frame, err := codec.NewPacketCodec().DecodePacket(buf.Bytes())
	if err != nil {
		return nil, sl427.WrapError(sl427.ErrCodeInvalidData, "解码失败", err)
	}

	// 7. 解析用户数据
	p, err := packet.ParseUserData(frame)
	if err != nil {
		return nil, sl427.WrapError(sl427.ErrCodeInvalidData, "解析失败", err)
	}

	// // 8. 更新统计信息
	// h.metrics.PacketsReceived++
	// h.metrics.LastReceiveTime = time.Now()

	h.logger.Printf("成功读取数据包: 长度=%d bytes", buf.Len())
	return p, nil
}

// SetLogger 实现Handler接口：设置日志接口
func (h *handlerImpl) SetLogger(logger types.Logger) {
	if logger != nil {
		h.logger = logger
	}
}

// Close 实现Handler接口：关闭处理器
func (h *handlerImpl) Close() error {
	return h.conn.Close()
}

// RemoteAddr 实现Handler接口：获取远程地址
func (h *handlerImpl) RemoteAddr() net.Addr {
	return h.conn.RemoteAddr()
}
