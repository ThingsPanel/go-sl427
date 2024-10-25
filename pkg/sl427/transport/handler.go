// pkg/sl427/transport/handler.go
package transport

import (
	"bufio"
	"encoding/binary"
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

func (h *handlerImpl) readPacket() (*packet.Packet, error) {
	// 1. 查找起始标识
	startByte, err := h.reader.ReadByte()
	if err != nil {
		return nil, sl427.WrapError(sl427.ErrCodeConnectionFailed, "读取起始字节失败", err)
	}

	for startByte != types.StartFlag {
		startByte, err = h.reader.ReadByte()
		if err != nil {
			return nil, sl427.WrapError(sl427.ErrCodeConnectionFailed, "查找起始标识失败", err)
		}
	}

	// 2. 读取地址和命令(5字节)
	headerBuf := make([]byte, 5)
	if _, err := io.ReadFull(h.reader, headerBuf); err != nil {
		return nil, sl427.WrapError(sl427.ErrCodeConnectionFailed, "读取地址和命令失败", err)
	}

	// 3. 读取长度字段(2字节)
	lengthBuf := make([]byte, 2)
	if _, err := io.ReadFull(h.reader, lengthBuf); err != nil {
		return nil, sl427.WrapError(sl427.ErrCodeConnectionFailed, "读取长度字段失败", err)
	}
	length := binary.BigEndian.Uint16(lengthBuf)

	// 4. 验证长度合理性
	if length < packet.MinPacketLen || length > uint16(h.config.MaxPacketSize) {
		return nil, sl427.WrapError(sl427.ErrCodeInvalidLength,
			fmt.Sprintf("无效的报文长度: %d", length), nil)
	}

	// 5. 创建完整的数据缓冲区并复制已读取的数据
	fullPacket := make([]byte, length)
	fullPacket[0] = startByte        // 起始标识
	copy(fullPacket[1:6], headerBuf) // 地址和命令
	copy(fullPacket[6:8], lengthBuf) // 长度字段

	// 6. 读取剩余数据(包括序列号、数据域、CRC和结束标识)
	remainingLength := int(length) - 8 // 减去已读取的字节
	if _, err := io.ReadFull(h.reader, fullPacket[8:length]); err != nil {
		return nil, sl427.WrapError(sl427.ErrCodeConnectionFailed,
			fmt.Sprintf("读取数据包剩余部分失败,期望%d字节", remainingLength), err)
	}

	// 7. 记录调试信息
	h.logger.Printf("接收到完整数据包: %X", fullPacket)

	// 8. 解析数据包
	p, err := packet.Parse(fullPacket)
	if err != nil {
		return nil, sl427.WrapError(sl427.ErrCodeInvalidData, "解析数据包失败", err)
	}

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
