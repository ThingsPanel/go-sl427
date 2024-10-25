// pkg/sl427/protocol/protocol.go
package protocol

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/ThingsPanel/go-sl427/pkg/sl427"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/codec"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/packet"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
)

// Protocol SL427协议接口定义
type Protocol interface {
	// ParseUploadData 解析上传数据报文
	ParseUploadData(pkt *packet.Packet) (*UploadData, error)

	// BuildUploadPacket 构建上传数据报文
	BuildUploadPacket(address uint32, data *UploadData) (*packet.Packet, error)

	// BuildHeartbeatPacket 构建心跳报文
	BuildHeartbeatPacket(address uint32) (*packet.Packet, error)

	// BuildResponsePacket 构建响应报文
	BuildResponsePacket(requestPkt *packet.Packet, success bool) (*packet.Packet, error)

	// Version 获取协议版本
	Version() string
}

// UploadData 上传数据结构
type UploadData struct {
	Timestamp time.Time  // 时间戳
	Items     []DataItem // 数据项列表
}

// DataItem 数据项
type DataItem struct {
	ID    uint16      // 数据项ID
	Type  byte        // 数据类型
	Value interface{} // 数据值
}

// ProtocolImpl 协议实现
type ProtocolImpl struct {
	packetCodec *codec.PacketCodec
	dataCodec   *codec.DataCodec
	version     string
}

// Config 协议配置
type Config struct {
	Version string       // 协议版本
	Logger  types.Logger // 日志接口
}

// Option 定义可选配置的函数类型
type Option func(*Config)

// WithVersion 设置协议版本
func WithVersion(version string) Option {
	return func(c *Config) {
		c.Version = version
	}
}

// WithLogger 设置日志接口
func WithLogger(logger types.Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// New 创建新的协议处理器
func New(opts ...Option) Protocol {
	// 默认配置
	config := &Config{
		Version: "SL427-2021",
		Logger:  types.DefaultLogger,
	}

	// 应用可选配置
	for _, opt := range opts {
		opt(config)
	}

	return &ProtocolImpl{
		packetCodec: codec.NewPacketCodec(),
		dataCodec:   codec.NewDataCodec(),
		version:     config.Version,
	}
}

// Version 获取协议版本
func (p *ProtocolImpl) Version() string {
	return p.version
}

// ParseUploadData 解析上传数据报文
func ParseUploadData(data []byte) (*UploadData, error) {
	if len(data) < types.TimestampLen+1 {
		return nil, sl427.WrapError(sl427.ErrCodeInvalidLength, "数据长度不足", nil)
	}

	// 解析时间戳
	timestamp, err := types.ParseTimeStamp(data[:types.TimestampLen])
	if err != nil {
		return nil, sl427.WrapError(sl427.ErrCodeInvalidFormat, "解析时间戳失败", err)
	}

	// 获取数据项数量
	itemCount := data[types.TimestampLen]
	offset := types.TimestampLen + 1

	// 解析数据项
	items := make([]DataItem, 0, itemCount)
	for i := byte(0); i < itemCount && offset < len(data); i++ {
		if offset+3 > len(data) {
			return nil, sl427.WrapError(sl427.ErrCodeInvalidData, fmt.Sprintf("数据项 %d 解析失败: 数据不足", i), nil)
		}

		// 读取ID和类型
		id := binary.BigEndian.Uint16(data[offset:])
		offset += 2
		dataType := data[offset]
		offset += 1

		// 根据类型解析值
		var value interface{}
		switch dataType {
		case types.TypeInt8:
			if offset+1 > len(data) {
				return nil, sl427.WrapError(sl427.ErrCodeInvalidData, fmt.Sprintf("数据项 %d Int8值解析失败: 数据不足", i), nil)
			}
			value = int8(data[offset])
			offset += 1

		case types.TypeInt16:
			if offset+2 > len(data) {
				return nil, sl427.WrapError(sl427.ErrCodeInvalidData, fmt.Sprintf("数据项 %d Int16值解析失败: 数据不足", i), nil)
			}
			value = int16(binary.BigEndian.Uint16(data[offset:]))
			offset += 2

		case types.TypeInt32:
			if offset+4 > len(data) {
				return nil, sl427.WrapError(sl427.ErrCodeInvalidData, fmt.Sprintf("数据项 %d Int32值解析失败: 数据不足", i), nil)
			}
			value = int32(binary.BigEndian.Uint32(data[offset:]))
			offset += 4

		case types.TypeString:
			if offset >= len(data) {
				return nil, sl427.WrapError(sl427.ErrCodeInvalidData, fmt.Sprintf("数据项 %d 字符串长度读取失败: 数据不足", i), nil)
			}
			strLen := data[offset]
			offset += 1
			if offset+int(strLen) > len(data) {
				return nil, sl427.WrapError(sl427.ErrCodeInvalidData, fmt.Sprintf("数据项 %d 字符串值读取失败: 数据不足", i), nil)
			}
			value = string(data[offset : offset+int(strLen)])
			offset += int(strLen)

		default:
			return nil, sl427.WrapError(sl427.ErrCodeInvalidType, fmt.Sprintf("数据项 %d 未知类型: %X", i, dataType), nil)
		}

		items = append(items, DataItem{
			ID:    id,
			Type:  dataType,
			Value: value,
		})
	}

	return &UploadData{
		Timestamp: timestamp.Time,
		Items:     items,
	}, nil
}

// EncodeUploadData 编码上传数据内容
func EncodeUploadData(data *UploadData) ([]byte, error) {
	// 预估缓冲区大小并编码数据
	bufSize := types.TimestampLen + 1 + len(data.Items)*10
	buf := make([]byte, 0, bufSize)

	// 编码时间戳
	timestamp := types.NewTimeStamp(data.Timestamp)
	buf = append(buf, timestamp.Bytes()...)

	// 写入数据项数量
	buf = append(buf, byte(len(data.Items)))

	// 编码每个数据项
	for _, item := range data.Items {
		// 写入ID
		idBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(idBytes, item.ID)
		buf = append(buf, idBytes...)

		// 写入类型
		buf = append(buf, item.Type)

		// 根据类型编码值
		switch item.Type {
		case types.TypeInt8:
			if v, ok := item.Value.(int8); ok {
				buf = append(buf, byte(v))
			} else {
				return nil, sl427.WrapError(sl427.ErrCodeInvalidType, fmt.Sprintf("数据项 %d 类型不匹配: 期望 Int8", item.ID), nil)
			}

		case types.TypeInt16:
			if v, ok := item.Value.(int16); ok {
				valBytes := make([]byte, 2)
				binary.BigEndian.PutUint16(valBytes, uint16(v))
				buf = append(buf, valBytes...)
			} else {
				return nil, sl427.WrapError(sl427.ErrCodeInvalidType, fmt.Sprintf("数据项 %d 类型不匹配: 期望 Int16", item.ID), nil)
			}

		case types.TypeInt32:
			if v, ok := item.Value.(int32); ok {
				valBytes := make([]byte, 4)
				binary.BigEndian.PutUint32(valBytes, uint32(v))
				buf = append(buf, valBytes...)
			} else {
				return nil, sl427.WrapError(sl427.ErrCodeInvalidType, fmt.Sprintf("数据项 %d 类型不匹配: 期望 Int32", item.ID), nil)
			}

		case types.TypeString:
			if v, ok := item.Value.(string); ok {
				if len(v) > 255 {
					return nil, sl427.WrapError(sl427.ErrCodeDataTooLong, fmt.Sprintf("数据项 %d 字符串过长", item.ID), nil)
				}
				buf = append(buf, byte(len(v)))
				buf = append(buf, v...)
			} else {
				return nil, sl427.WrapError(sl427.ErrCodeInvalidType, fmt.Sprintf("数据项 %d 类型不匹配: 期望 String", item.ID), nil)
			}

		default:
			return nil, sl427.WrapError(sl427.ErrCodeInvalidType, fmt.Sprintf("数据项 %d 未知类型: %X", item.ID, item.Type), nil)
		}
	}

	return buf, nil
}

// ParseUploadData 实现Protocol接口：解析上传数据报文
func (p *ProtocolImpl) ParseUploadData(pkt *packet.Packet) (*UploadData, error) {
	if pkt.Header.Command != types.CmdUpload {
		return nil, sl427.WrapError(sl427.ErrCodeInvalidData, "非上传数据报文", fmt.Errorf("command: %X", pkt.Header.Command))
	}
	return ParseUploadData(pkt.Data)
}

// BuildUploadPacket 实现Protocol接口：构建上传数据报文
func (p *ProtocolImpl) BuildUploadPacket(address uint32, data *UploadData) (*packet.Packet, error) {
	// 编码数据
	dataBytes, err := EncodeUploadData(data)
	if err != nil {
		return nil, sl427.WrapError(sl427.ErrCodeInvalidData, "编码数据失败", err)
	}

	// 构建报文
	return packet.NewPacket(address, types.CmdUpload, dataBytes)
}

// BuildHeartbeatPacket 实现Protocol接口：构建心跳报文
func (p *ProtocolImpl) BuildHeartbeatPacket(address uint32) (*packet.Packet, error) {
	return packet.NewPacket(address, types.CmdHeartbeat, nil)
}

// BuildResponsePacket 实现Protocol接口：构建响应报文
func (p *ProtocolImpl) BuildResponsePacket(requestPkt *packet.Packet, success bool) (*packet.Packet, error) {
	status := types.RespSuccess
	if !success {
		status = types.RespError
	}
	return packet.NewPacket(requestPkt.Header.Address, requestPkt.Header.Command, []byte{status})
}
