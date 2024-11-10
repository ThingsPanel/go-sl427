// pkg/sl427/codec/packet_codec.go
package codec

import (
	"bytes"
	"fmt"

	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
)

// PacketCodec 报文编解码器
type PacketCodec struct{}

// NewPacketCodec 创建新的编解码器实例
func NewPacketCodec() *PacketCodec {
	return &PacketCodec{}
}

// DecodePacket 将字节流解码为Frame
func (c *PacketCodec) DecodePacket(data []byte) (*types.Frame, error) {
	// 1. 基本长度检查
	if len(data) < types.MinFrameLen {
		return nil, fmt.Errorf("packet too short: %d", len(data))
	}

	// 2. 检查起始和结束标识
	if data[0] != types.StartFlag || data[2] != types.StartFlag {
		return nil, fmt.Errorf("invalid start flag")
	}
	if data[len(data)-1] != types.EndFlag {
		return nil, fmt.Errorf("invalid end flag")
	}

	// 3. 获取用户数据区长度
	length := data[1]
	expectedLen := int(length) + 5 // 帧头(3) + CS(1) + 结束符(1)
	if len(data) != expectedLen {
		return nil, fmt.Errorf("invalid packet length")
	}

	// 4. 提取用户数据区
	userDataStart := 3
	userDataEnd := len(data) - 2
	userData := data[userDataStart:userDataEnd]

	// 5. 校验CS
	expectedCS := c.calculateCS(userData)
	actualCS := data[len(data)-2]
	if expectedCS != actualCS {
		return nil, fmt.Errorf("CS 校验失败，期望 %X, 实际 %X", expectedCS, actualCS)
	}

	// 6. 构建Frame对象
	frame := &types.Frame{
		Head: types.Header{
			StartFlag1: data[0],
			Length:     length,
			StartFlag2: data[2],
		},
		UserDataRaw: userData,
		CS:          actualCS,
		EndFlag:     data[len(data)-1],
	}

	return frame, nil
}

// EncodePacket 将Frame编码为字节流
func (c *PacketCodec) EncodePacket(frame *types.Frame) ([]byte, error) {
	// 预分配缓冲区
	buf := bytes.Buffer{}

	// 1. 写入帧头
	buf.WriteByte(frame.Head.StartFlag1)
	buf.WriteByte(frame.Head.Length)
	buf.WriteByte(frame.Head.StartFlag2)

	// 2. 写入用户数据区
	buf.Write(frame.UserDataRaw)

	// 3. 计算并写入CS
	cs := c.calculateCS(frame.UserDataRaw)
	buf.WriteByte(cs)

	// 4. 写入帧结束标识
	buf.WriteByte(types.EndFlag)

	return buf.Bytes(), nil
}

// calculateCS 计算用户数据区的CRC校验
// 生成多项式: X7+X6+X5+X2+1 = 1110 0100
func (c *PacketCodec) calculateCS(data []byte) byte {
	var crc byte
	const poly = 0xE4 // 生成多项式: X7+X6+X5+X2+1 = 1110 0100

	for _, b := range data {
		crc ^= b // 与输入字节异或

		for i := 0; i < 8; i++ {
			if (crc & 0x80) != 0 { // 检查最高位是1
				crc = (crc << 1) ^ poly // 左移并异或多项式
			} else {
				crc = crc << 1 // 只左移
			}
		}
	}

	return crc & 0x7F // 返回低7位作为校验值
}
