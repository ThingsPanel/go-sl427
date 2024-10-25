// pkg/sl427/packet/packet.go
package packet

import (
	"encoding/binary"
	"fmt"
)

const (
	StartFlag byte = 0x68
	EndFlag   byte = 0x16

	HeaderLen    = 9 // 固定头部长度(起始标识1 + 地址4 + 命令码1 + 长度2 + 序列号1)
	ChecksumLen  = 2 // CRC校验码长度
	EndFlagLen   = 1 // 结束标识长度
	MinPacketLen = HeaderLen + ChecksumLen + EndFlagLen
	MaxPacketLen = 1024
)

// Packet 完整报文结构
type Packet struct {
	Header *Header
	Data   []byte
	CRC    uint16
}

// Header 报文头结构
type Header struct {
	StartFlag byte   // 起始标识
	Address   uint32 // 地址域
	Command   byte   // 命令码
	Length    uint16 // 总长度
	SerialNum byte   // 流水号
}

// NewPacket 创建新的报文
func NewPacket(address uint32, command byte, data []byte) (*Packet, error) {
	dataLen := len(data)
	if dataLen > MaxPacketLen-MinPacketLen {
		return nil, fmt.Errorf("数据长度超出限制: %d > %d", dataLen, MaxPacketLen-MinPacketLen)
	}

	// 计算总长度：头部 + 数据 + CRC + 结束标识
	totalLen := HeaderLen + dataLen + ChecksumLen + EndFlagLen

	header := &Header{
		StartFlag: StartFlag,
		Address:   address,
		Command:   command,
		Length:    uint16(totalLen),
		SerialNum: 0,
	}

	p := &Packet{
		Header: header,
		Data:   data,
	}

	// 计算CRC
	p.CRC = p.CalculateChecksum()

	return p, nil
}

// Bytes 将报文转换为字节数组
func (p *Packet) Bytes() []byte {
	totalLen := HeaderLen + len(p.Data) + ChecksumLen + EndFlagLen
	buf := make([]byte, totalLen)

	// 1. 写入头部
	buf[0] = p.Header.StartFlag
	binary.BigEndian.PutUint32(buf[1:5], p.Header.Address)
	buf[5] = p.Header.Command
	binary.BigEndian.PutUint16(buf[6:8], uint16(totalLen))
	buf[8] = p.Header.SerialNum

	// 2. 写入数据
	if len(p.Data) > 0 {
		copy(buf[HeaderLen:], p.Data)
	}

	// 3. 计算并写入CRC
	p.CRC = p.CalculateChecksum()
	binary.BigEndian.PutUint16(buf[totalLen-3:totalLen-1], p.CRC)

	// 4. 写入结束标识
	buf[totalLen-1] = EndFlag

	return buf
}

// CalculateChecksum 计算校验和
func (p *Packet) CalculateChecksum() uint16 {
	// 计算长度：头部 + 数据
	length := HeaderLen + len(p.Data)
	data := make([]byte, length)

	// 复制头部
	data[0] = p.Header.StartFlag
	binary.BigEndian.PutUint32(data[1:5], p.Header.Address)
	data[5] = p.Header.Command
	binary.BigEndian.PutUint16(data[6:8], p.Header.Length)
	data[8] = p.Header.SerialNum

	// 复制数据
	if len(p.Data) > 0 {
		copy(data[HeaderLen:], p.Data)
	}

	// 计算校验和
	var sum uint16
	for _, b := range data {
		sum += uint16(b)
	}
	return sum
}

// Parse 解析报文
func Parse(data []byte) (*Packet, error) {
	// 1. 基本长度检查
	if len(data) < MinPacketLen {
		return nil, fmt.Errorf("数据长度(%d)小于最小长度(%d)", len(data), MinPacketLen)
	}

	// 2. 验证起始标识
	if data[0] != StartFlag {
		return nil, fmt.Errorf("无效的起始标识: 0x%02X", data[0])
	}

	// 3. 解析头部
	header := &Header{
		StartFlag: data[0],
		Address:   binary.BigEndian.Uint32(data[1:5]),
		Command:   data[5],
		Length:    binary.BigEndian.Uint16(data[6:8]),
		SerialNum: data[8],
	}

	// 4. 验证长度
	if header.Length != uint16(len(data)) {
		return nil, fmt.Errorf("数据长度不匹配: 报文声明 %d, 实际长度 %d", header.Length, len(data))
	}

	// 5. 提取数据域
	dataLen := len(data) - MinPacketLen
	var packetData []byte
	if dataLen > 0 {
		packetData = make([]byte, dataLen)
		copy(packetData, data[HeaderLen:HeaderLen+dataLen])
	}

	// 6. 提取CRC和结束标识
	crc := binary.BigEndian.Uint16(data[len(data)-3 : len(data)-1])
	endFlag := data[len(data)-1]

	// 7. 验证结束标识
	if endFlag != EndFlag {
		return nil, fmt.Errorf("无效的结束标识: 0x%02X", endFlag)
	}

	packet := &Packet{
		Header: header,
		Data:   packetData,
		CRC:    crc,
	}

	// 8. 验证CRC
	calculatedCRC := packet.CalculateChecksum()
	if calculatedCRC != crc {
		return nil, fmt.Errorf("CRC校验失败: 计算值=0x%04X, 期望值=0x%04X", calculatedCRC, crc)
	}

	return packet, nil
}
