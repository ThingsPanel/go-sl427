// pkg/sl427/codec/codec.go
package codec

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/ThingsPanel/go-sl427/pkg/sl427"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/packet"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
)

var (
	ErrInvalidType = errors.New("invalid data type")
	ErrDataTooLong = errors.New("data too long")
)

// Codec 编解码器接口
type Codec interface {
	// Encode 将数据编码为字节流
	Encode(interface{}) ([]byte, error)

	// Decode 将字节流解码为数据
	Decode([]byte) (interface{}, error)
}

// DataCodec 数据编解码器
type DataCodec struct{}

// NewDataCodec 创建数据编解码器
func NewDataCodec() *DataCodec {
	return &DataCodec{}
}

// EncodeInt 编码整数
func (c *DataCodec) EncodeInt(value int, size int) ([]byte, error) {
	buf := make([]byte, size)
	switch size {
	case 1:
		buf[0] = byte(value)
	case 2:
		binary.BigEndian.PutUint16(buf, uint16(value))
	case 4:
		binary.BigEndian.PutUint32(buf, uint32(value))
	default:
		return nil, ErrInvalidType
	}
	return buf, nil
}

// DecodeInt 解码整数
func (c *DataCodec) DecodeInt(data []byte, size int) (int, error) {
	if len(data) < size {
		return 0, sl427.ErrInvalidLength
	}

	switch size {
	case 1:
		return int(data[0]), nil
	case 2:
		return int(binary.BigEndian.Uint16(data)), nil
	case 4:
		return int(binary.BigEndian.Uint32(data)), nil
	default:
		return 0, ErrInvalidType
	}
}

// EncodeString 编码字符串
func (c *DataCodec) EncodeString(value string, maxLen int) ([]byte, error) {
	if len(value) > maxLen {
		return nil, ErrDataTooLong
	}
	return []byte(value), nil
}

// DecodeString 解码字符串
func (c *DataCodec) DecodeString(data []byte, maxLen int) (string, error) {
	if len(data) > maxLen {
		return "", ErrDataTooLong
	}
	return string(data), nil
}

// EncodeTime 编码时间
func (c *DataCodec) EncodeTime(t types.TimeStamp) ([]byte, error) {
	return t.Bytes(), nil
}

// DecodeTime 解码时间
func (c *DataCodec) DecodeTime(data []byte) (types.TimeStamp, error) {
	return types.ParseTimeStamp(data)
}

// PacketCodec 报文编解码器
type PacketCodec struct {
	dataCodec *DataCodec
}

// NewPacketCodec 创建报文编解码器
func NewPacketCodec() *PacketCodec {
	return &PacketCodec{
		dataCodec: NewDataCodec(),
	}
}

func (c *PacketCodec) EncodePacket(p *packet.Packet) ([]byte, error) {
	// 1. 计算实际数据长度(不包括CRC和EndFlag)
	dataLen := len(p.Data)
	totalLen := packet.MinPacketLen + dataLen

	// 2. 创建缓冲区(加上CRC和EndFlag的长度)
	buf := make([]byte, totalLen+3)

	// 3. 编码报文头
	buf[0] = p.Header.StartFlag
	binary.BigEndian.PutUint32(buf[1:5], p.Header.Address)
	buf[5] = p.Header.Command
	binary.BigEndian.PutUint16(buf[6:8], uint16(totalLen))
	buf[8] = p.Header.SerialNum

	// 4. 写入数据域(如果有)
	if dataLen > 0 {
		copy(buf[9:9+dataLen], p.Data)
	}

	// 5. 计算并写入CRC (计算除CRC和EndFlag外的所有数据)
	crc := calculateCRC(buf[:totalLen])
	binary.BigEndian.PutUint16(buf[totalLen:totalLen+2], crc)

	// 6. 写入结束标识
	buf[totalLen+2] = packet.EndFlag

	return buf, nil
}

func (c *PacketCodec) DecodePacket(data []byte) (*packet.Packet, error) {
	// 1. 基本长度检查
	if len(data) < packet.MinPacketLen {
		return nil, fmt.Errorf("数据长度(%d)小于最小长度(%d)", len(data), packet.MinPacketLen)
	}

	// 2. 检查起始标识
	if data[0] != packet.StartFlag {
		return nil, fmt.Errorf("无效的起始标识: %X", data[0])
	}

	// 3. 获取声明的长度
	declaredLen := binary.BigEndian.Uint16(data[6:8])

	// 4. 验证实际长度(包括CRC和EndFlag)
	if int(declaredLen)+3 != len(data) {
		return nil, fmt.Errorf("数据长度不匹配: 声明 %d, 实际 %d", declaredLen, len(data)-3)
	}

	// 5. 创建Packet对象
	p := &packet.Packet{
		Header: &packet.Header{
			StartFlag: data[0],
			Address:   binary.BigEndian.Uint32(data[1:5]),
			Command:   data[5],
			Length:    declaredLen,
			SerialNum: data[8],
		},
	}

	// 6. 提取数据域
	dataLen := int(declaredLen) - packet.MinPacketLen
	if dataLen > 0 {
		p.Data = make([]byte, dataLen)
		copy(p.Data, data[9:9+dataLen])
	}

	// 7. 验证CRC
	p.CRC = binary.BigEndian.Uint16(data[len(data)-3 : len(data)-1])
	calculatedCRC := calculateCRC(data[:len(data)-3])
	if calculatedCRC != p.CRC {
		return nil, fmt.Errorf("CRC校验失败: 计算值 %X, 期望值 %X", calculatedCRC, p.CRC)
	}

	// 8. 验证结束标识
	if data[len(data)-1] != packet.EndFlag {
		return nil, fmt.Errorf("无效的结束标识: %X", data[len(data)-1])
	}

	return p, nil
}

// calculateCRC 计算CRC校验和
func calculateCRC(data []byte) uint16 {
	var sum uint16
	for _, b := range data {
		sum += uint16(b)
	}
	return sum
}
