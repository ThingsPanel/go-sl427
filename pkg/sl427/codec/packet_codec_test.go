package codec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPacketCodec_Simple(t *testing.T) {
	codec := NewPacketCodec()

	// 构造用户数据区
	userData := []byte{
		0x80,                         // 控制域
		0x01, 0x02, 0x03, 0x04, 0x05, // 地址域(5字节)
		0xC0, // 功能码
		0x01, // 数据域(1字节)
	}

	// 计算CS
	cs := calculateCS(userData)

	// 构造完整帧
	packet := []byte{
		0x68, // 起始符1
		0x08, // L=8(用户数据区长度)
		0x68, // 起始符2
	}
	packet = append(packet, userData...) // 用户数据区
	packet = append(packet, cs)          // CS
	packet = append(packet, 0x16)        // 结束符

	// 1. 测试解码
	frame, err := codec.DecodePacket(packet)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证解码结果
	if frame.Head.Length != 0x08 {
		t.Errorf("长度字段错误: want 0x08, got 0x%02x", frame.Head.Length)
	}

	// 2. 测试编码
	encoded, err := codec.EncodePacket(frame)
	if err != nil {
		t.Fatalf("编码失败: %v", err)
	}

	// 3. 验证长度
	expectedLen := int(frame.Head.Length) + 5
	if len(encoded) != expectedLen {
		t.Errorf("编码后长度错误: want %d, got %d", expectedLen, len(encoded))
	}
}

// 用于测试的CS计算函数，与codec中的实现保持一致
func calculateCS(data []byte) byte {
	var cs byte
	for _, b := range data {
		cs ^= b
	}
	return cs
}

func TestPacketCodec_DecodeInvalid(t *testing.T) {
	codec := NewPacketCodec()

	// 测试长度太短
	_, err := codec.DecodePacket([]byte{0x68, 0x01})
	if err == nil {
		t.Error("应该检测出数据太短")
	}
}

func TestPacketCodec_InvalidInput(t *testing.T) {
	codec := NewPacketCodec()

	t.Run("decode too short packet", func(t *testing.T) {
		_, err := codec.DecodePacket([]byte{0x68, 0x01})
		assert.Error(t, err)
	})

	t.Run("decode invalid start flag", func(t *testing.T) {
		_, err := codec.DecodePacket([]byte{0x00, 0x01, 0x68, 0x01, 0x02, 0x03, 0x16})
		assert.Error(t, err)
	})

	t.Run("decode invalid end flag", func(t *testing.T) {
		_, err := codec.DecodePacket([]byte{0x68, 0x01, 0x68, 0x01, 0x02, 0x03, 0x00})
		assert.Error(t, err)
	})
}
