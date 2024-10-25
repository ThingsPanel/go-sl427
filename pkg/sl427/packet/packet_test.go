// pkg/sl427/packet/packet_test.go
package packet

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPacket_Basic(t *testing.T) {
	// 测试创建新数据包
	t.Run("NewPacket", func(t *testing.T) {
		p, err := NewPacket(0x12345678, 0x02, []byte{0x01, 0x02, 0x03})
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, byte(0x68), p.Header.StartFlag)
		assert.Equal(t, uint32(0x12345678), p.Header.Address)
		assert.Equal(t, byte(0x02), p.Header.Command)
	})

	// 测试数据包长度限制
	t.Run("PacketLengthLimit", func(t *testing.T) {
		// 创建超长数据
		data := make([]byte, MaxPacketLen)
		_, err := NewPacket(0x01, 0x02, data)
		assert.Error(t, err)
	})

	// 测试最小数据包
	t.Run("MinimalPacket", func(t *testing.T) {
		p, err := NewPacket(0x01, 0x03, nil)
		assert.NoError(t, err)
		assert.Equal(t, MinPacketLen, int(p.Header.Length))
	})
}

func TestPacket_Conversion(t *testing.T) {
	// 测试数据包与字节流转换
	t.Run("PacketToBytes", func(t *testing.T) {
		p, err := NewPacket(0x12345678, 0x02, []byte{0x01, 0x02, 0x03})
		assert.NoError(t, err)

		data := p.Bytes()
		assert.Equal(t, byte(0x68), data[0])           // 起始标识
		assert.Equal(t, byte(0x16), data[len(data)-1]) // 结束标识
	})

	// 测试CRC计算
	t.Run("CRCCalculation", func(t *testing.T) {
		p, _ := NewPacket(0x12345678, 0x02, []byte{0x01, 0x02, 0x03})
		data := p.Bytes()
		parsed, err := Parse(data)
		assert.NoError(t, err)
		assert.Equal(t, p.CRC, parsed.CRC)
	})
}

func TestPacket_Parse(t *testing.T) {
	// 正常数据包解析
	// 正常数据包解析
	t.Run("ValidPacket", func(t *testing.T) {
		// 构造心跳包数据
		packet := []byte{
			0x68,                   // 起始标识
			0x12, 0x34, 0x56, 0x78, // 地址
			0x03,       // 命令(心跳)
			0x00, 0x18, // 长度(24字节)
			0x01, // 序号
			// 时间戳(YYMMDDhhmmss)
			0x32, 0x34, 0x31, 0x30, 0x32, 0x35,
			0x32, 0x32, 0x33, 0x37, 0x30, 0x31,
		}

		// 计算CRC
		var sum uint16
		for _, b := range packet {
			sum += uint16(b)
		}

		// 添加CRC和结束标识
		crcBytes := []byte{byte(sum >> 8), byte(sum)}
		packet = append(packet, crcBytes...)
		packet = append(packet, 0x16) // 结束标识

		t.Logf("测试数据包: % X", packet)
		t.Logf("计算得到的CRC: %04X", sum)

		// 解析数据包
		p, err := Parse(packet)
		assert.NoError(t, err, "解析数据包失败")
		assert.NotNil(t, p, "解析结果不应为空")

		// 验证头部字段
		assert.Equal(t, byte(0x68), p.Header.StartFlag, "起始标识不匹配")
		assert.Equal(t, uint32(0x12345678), p.Header.Address, "地址不匹配")
		assert.Equal(t, byte(0x03), p.Header.Command, "命令码不匹配")
		assert.Equal(t, uint16(0x0018), p.Header.Length, "长度不匹配")
		assert.Equal(t, byte(0x01), p.Header.SerialNum, "序列号不匹配")

		// 验证数据长度
		assert.Equal(t, 12, len(p.Data), "数据域长度应为12字节")

		// 验证CRC
		assert.Equal(t, sum, p.CRC, "CRC不匹配")

		// 将解析后的数据包重新编码
		encoded := p.Bytes()
		assert.Equal(t, packet, encoded, "重新编码后的数据应与原始数据相同")
	})

	// 验证CRC校验失败的情况
	t.Run("InvalidCRC", func(t *testing.T) {
		invalidPacket := []byte{
			0x68,                   // 起始标识
			0x12, 0x34, 0x56, 0x78, // 地址
			0x03,       // 命令(心跳)
			0x00, 0x18, // 长度(24字节)
			0x01, // 序号
			// 时间戳(YYMMDDhhmmss)
			0x32, 0x34, 0x31, 0x30, 0x32, 0x35,
			0x32, 0x32, 0x33, 0x37, 0x30, 0x31,
			0xFF, 0xFF, // 错误的CRC
			0x16, // 结束标识
		}

		_, err := Parse(invalidPacket)
		assert.Error(t, err, "应检测出CRC校验错误")
		assert.Contains(t, err.Error(), "CRC校验失败", "错误消息应包含CRC校验失败说明")
	})

	// 测试长度字段不匹配
	t.Run("LengthMismatch", func(t *testing.T) {
		invalidPacket := []byte{
			0x68,                   // 起始标识
			0x12, 0x34, 0x56, 0x78, // 地址
			0x03,       // 命令(心跳)
			0x00, 0x20, // 错误的长度(32字节)
			0x01, // 序号
			// 时间戳(YYMMDDhhmmss)
			0x32, 0x34, 0x31, 0x30, 0x32, 0x35,
			0x32, 0x32, 0x33, 0x37, 0x30, 0x31,
			0x00, 0x00, // CRC
			0x16, // 结束标识
		}

		_, err := Parse(invalidPacket)
		assert.Error(t, err, "应检测出长度不匹配错误")
		assert.Contains(t, err.Error(), "数据长度不匹配", "错误消息应包含长度不匹配说明")
	})

	// 测试长度字段与实际长度不匹配的情况
	t.Run("InvalidLength", func(t *testing.T) {
		// 构造一个长度声明与实际不符的数据包
		packet := []byte{
			0x68,                   // 起始标识
			0x12, 0x34, 0x56, 0x78, // 地址
			0x03,       // 命令(心跳)
			0x00, 0x18, // 长度声明(24字节)，但实际数据不足
			0x01, // 序号
			// 时间戳(只有10字节，少了2字节)
			0x32, 0x34, 0x31, 0x30, 0x32, 0x35, 0x32, 0x32, 0x33, 0x37,
			0x0A, 0xBC, // CRC
			0x16, // 结束标识
		}

		// 记录构造的数据包内容
		t.Logf("测试数据包(长度不匹配): % X", packet)
		t.Logf("声明长度: %d, 实际长度: %d", 0x18, len(packet))

		// 验证解析结果
		_, err := Parse(packet)
		assert.Error(t, err, "应当检测出长度不匹配错误")
		assert.Contains(t, err.Error(), "数据长度不匹配", "错误消息应当包含长度不匹配说明")
	})

	// 测试数据包过短的情况
	t.Run("PacketTooShort", func(t *testing.T) {
		// 构造一个不完整的数据包
		shortPacket := []byte{
			0x68,                   // 起始标识
			0x12, 0x34, 0x56, 0x78, // 地址
			0x03, // 命令
		}

		t.Logf("测试数据包(过短): % X", shortPacket)

		_, err := Parse(shortPacket)
		assert.Error(t, err, "应当检测出数据包过短错误")
		assert.Contains(t, err.Error(), "数据长度", "错误消息应当说明长度问题")
	})

	// 测试长度字段异常
	t.Run("InvalidLength", func(t *testing.T) {
		packet := []byte{
			0x68,                   // 起始标识
			0x12, 0x34, 0x56, 0x78, // 地址
			0x03,       // 命令(心跳)
			0x00, 0x20, // 长度声明(32字节，实际数据小于这个值)
			0x01, // 序号
			// 时间戳
			0x32, 0x34, 0x31, 0x30, 0x32, 0x35,
			0x0A, 0xBC, // CRC
			0x16, // 结束标识
		}

		// 记录日志便于调试
		t.Logf("测试数据包(长度不匹配): % X", packet)

		// 验证解析结果
		_, err := Parse(packet)
		assert.Error(t, err, "应当检测出长度不匹配错误")
		assert.Contains(t, err.Error(), "数据长度不匹配", "错误消息应当包含长度不匹配说明")
	})

	// 补充测试用例 - CRC校验
	t.Run("CRCVerification", func(t *testing.T) {
		packetHex := "681234567803001801323431303235323233373031FFFF16" // 错误的CRC
		data, err := hex.DecodeString(packetHex)
		assert.NoError(t, err, "解码十六进制字符串失败")

		_, err = Parse(data)
		assert.Error(t, err, "应当检测出CRC校验错误")
		assert.Contains(t, err.Error(), "CRC校验失败", "错误消息应当包含CRC校验失败说明")
	})

	// 测试无效的起始标识
	t.Run("InvalidStartFlag", func(t *testing.T) {
		// 构造一个除起始标识外都正确的数据包
		packet := []byte{
			0x00,                   // 错误的起始标识(应该是0x68)
			0x12, 0x34, 0x56, 0x78, // 地址
			0x03,       // 命令(心跳)
			0x00, 0x18, // 长度
			0x01, // 序号
			// 时间戳
			0x32, 0x34, 0x31, 0x30, 0x32, 0x35,
			0x32, 0x32, 0x33, 0x37, 0x30, 0x31,
			0x0A, 0xBC, // CRC
			0x16, // 结束标识
		}

		_, err := Parse(packet)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的起始标识")
	})

	// 测试实际上传数据包
	t.Run("UploadDataPacket", func(t *testing.T) {
		// 模拟一个实际的上传数据包
		// 68 12345678 02 003A 01 时间戳+数据项
		packetHex := "681234567802003A013234313032353232333730310503E9030000303903EA030000162E03EB0202D303EC02099803ED046E6F726D616C0D7B16"
		data, _ := hex.DecodeString(packetHex)

		p, err := Parse(data)
		assert.NoError(t, err)
		assert.Equal(t, uint32(0x12345678), p.Header.Address)
		assert.Equal(t, byte(0x02), p.Header.Command) // 上传命令
		assert.Equal(t, uint16(0x3A), p.Header.Length)
		assert.Equal(t, byte(0x01), p.Header.SerialNum)

		// 验证数据域前12字节是时间戳
		assert.Equal(t, 12, len(p.Data[:12]))
		// 验证数据项数量
		assert.Equal(t, byte(0x05), p.Data[12])
	})
}

func TestPacket_RealWorldExamples(t *testing.T) {
	// 测试实际观察到的数据包案例
	cases := []struct {
		name    string
		hexData string
		wantErr bool
	}{
		{
			name:    "HeartbeatPacket",
			hexData: "68123456780300180232343130323532323332313303F416",
			wantErr: false,
		},
		{
			name:    "UploadDataPacket",
			hexData: "681234567802003A013234313032353232333730310503E9030000303903EA030000162E03EB0202D303EC02099803ED046E6F726D616C0D7B16",
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := hex.DecodeString(tc.hexData)
			assert.NoError(t, err)

			p, err := Parse(data)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// 重新编码并比较
			encoded := p.Bytes()
			assert.Equal(t, data, encoded)
		})
	}
}

func TestPacket_CalculateChecksum(t *testing.T) {
	t.Run("ChecksumCalculation", func(t *testing.T) {
		// 创建一个测试数据包
		p, err := NewPacket(0x12345678, 0x02, []byte{0x01, 0x02, 0x03})
		assert.NoError(t, err)

		// 计算校验和
		crc := p.CalculateChecksum()

		// 重新计算并比较
		data := p.Bytes()
		parsed, err := Parse(data)
		assert.NoError(t, err)
		assert.Equal(t, crc, parsed.CRC)
	})

	t.Run("ChecksumVerification", func(t *testing.T) {
		packetHex := "681234567802003A013234313032353232333730310503E9030000303903EA030000162E03EB0202D303EC02099803ED046E6F726D616C0D7B16"
		data, _ := hex.DecodeString(packetHex)

		p, err := Parse(data)
		assert.NoError(t, err)

		// 验证CRC是否正确
		calculatedCRC := p.CalculateChecksum()
		assert.Equal(t, p.CRC, calculatedCRC)
	})
}
