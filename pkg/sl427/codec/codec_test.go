package codec

import (
	"sync"
	"testing"
	"time"

	"github.com/ThingsPanel/go-sl427/pkg/sl427/packet"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
	"github.com/stretchr/testify/assert"
)

// TestDataCodec 测试数据编解码器
func TestDataCodec(t *testing.T) {
	codec := NewDataCodec()

	// 测试整数编解码
	t.Run("IntegerCodec", func(t *testing.T) {
		tests := []struct {
			name  string
			value int
			size  int
		}{
			{"Int8", 123, 1},
			{"Int16", 12345, 2},
			{"Int32", 123456789, 4},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				encoded, err := codec.EncodeInt(tt.value, tt.size)
				assert.NoError(t, err)
				assert.Equal(t, tt.size, len(encoded))

				decoded, err := codec.DecodeInt(encoded, tt.size)
				assert.NoError(t, err)
				assert.Equal(t, tt.value, decoded)
			})
		}
	})

	// 测试字符串编解码
	t.Run("StringCodec", func(t *testing.T) {
		tests := []struct {
			name      string
			value     string
			maxLen    int
			expectErr bool
		}{
			{"Normal", "TestString", 255, false},
			{"Empty", "", 255, false},
			{"MaxLength", "A", 1, false},
			{"TooLong", "TestString", 4, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				encoded, err := codec.EncodeString(tt.value, tt.maxLen)
				if tt.expectErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)

				decoded, err := codec.DecodeString(encoded, tt.maxLen)
				assert.NoError(t, err)
				assert.Equal(t, tt.value, decoded)
			})
		}
	})

	// 测试时间编解码
	t.Run("TimeCodec", func(t *testing.T) {
		tests := []struct {
			name string
			time time.Time
		}{
			{"CurrentTime", time.Now()},
			{"PastTime", time.Date(2023, 1, 1, 12, 0, 0, 0, time.Local)},
			{"FutureTime", time.Now().Add(24 * time.Hour)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ts := types.NewTimeStamp(tt.time)
				encoded, err := codec.EncodeTime(ts)
				assert.NoError(t, err)

				decoded, err := codec.DecodeTime(encoded)
				assert.NoError(t, err)
				assert.Equal(t, ts.Format("2006-01-02 15:04:05"),
					decoded.Format("2006-01-02 15:04:05"))
			})
		}
	})
}

func TestPacketCodec(t *testing.T) {
	packetCodec := NewPacketCodec()

	t.Run("FullPacket", func(t *testing.T) {
		// 创建一个固定的测试数据包
		testData := []byte{0x01, 0x02, 0x03}
		original, err := packet.NewPacket(
			0x12345678,      // address
			types.CmdUpload, // command
			testData,        // data
		)
		assert.NoError(t, err)
		original.Header.SerialNum = 0x01

		// 编码数据包
		encoded, err := packetCodec.EncodePacket(original)
		assert.NoError(t, err)

		// 解码数据包
		decoded, err := packetCodec.DecodePacket(encoded)
		assert.NoError(t, err)

		// 验证字段
		assert.Equal(t, packet.StartFlag, decoded.Header.StartFlag)
		assert.Equal(t, original.Header.Address, decoded.Header.Address)
		assert.Equal(t, original.Header.Command, decoded.Header.Command)
		assert.Equal(t, original.Header.SerialNum, decoded.Header.SerialNum)
		assert.Equal(t, packet.MinPacketLen+len(testData), int(decoded.Header.Length))
		assert.Equal(t, testData, decoded.Data)

		// 计算期望的CRC
		expectedCRC := uint16(0)
		for _, b := range encoded[:len(encoded)-3] {
			expectedCRC += uint16(b)
		}
		assert.Equal(t, expectedCRC, decoded.CRC)
	})
}

// TestConcurrentUsage 测试并发使用
func TestConcurrentUsage(t *testing.T) {
	packetCodec := NewPacketCodec()
	var wg sync.WaitGroup

	testFunc := func() {
		defer wg.Done()
		p, err := packet.NewPacket(0x01, types.CmdHeartbeat, nil)
		assert.NoError(t, err)

		encoded, err := packetCodec.EncodePacket(p)
		assert.NoError(t, err)

		decoded, err := packetCodec.DecodePacket(encoded)
		assert.NoError(t, err)
		assert.Equal(t, p.Header.Command, decoded.Header.Command)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go testFunc()
	}

	wg.Wait()
}

// TestErrorCases 测试错误场景
func TestErrorCases(t *testing.T) {
	packetCodec := NewPacketCodec()

	t.Run("InvalidPacket", func(t *testing.T) {
		tests := []struct {
			name      string
			data      []byte
			expectErr bool
		}{
			{"TooShort", []byte{0x68, 0x01}, true},
			{"InvalidStartFlag", []byte{0x00, 0x01, 0x02, 0x03}, true},
			{"InvalidEndFlag", []byte{
				0x68, 0x00, 0x00, 0x00, 0x01,
				0x02, 0x00, 0x0C, 0x01,
				0x00, 0x00,
				0x00, // Wrong end flag
			}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := packetCodec.DecodePacket(tt.data)
				if tt.expectErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

// TestCompleteDataFlow 测试完整数据流
func TestCompleteDataFlow(t *testing.T) {
	packetCodec := NewPacketCodec()
	dataCodec := NewDataCodec()

	// 准备测试数据
	testData := []struct {
		id    uint16
		value int32
		size  int
	}{
		{1001, int32(12345), 4},
		{1002, int32(5678), 4},
		{1003, int32(723), 4},
	}

	// 构建数据域
	payload := make([]byte, 0, 32)

	// 添加时间戳
	ts := types.NewTimeStamp(time.Now())
	tsBytes, err := dataCodec.EncodeTime(ts)
	assert.NoError(t, err)
	payload = append(payload, tsBytes...)

	// 添加数据项数量
	payload = append(payload, byte(len(testData)))

	// 编码数据项
	for _, d := range testData {
		// 添加ID
		idBytes := make([]byte, 2)
		idBytes[0] = byte(d.id >> 8)
		idBytes[1] = byte(d.id)
		payload = append(payload, idBytes...)

		// 添加类型
		payload = append(payload, types.TypeInt32)

		// 编码值
		valueBytes, err := dataCodec.EncodeInt(int(d.value), d.size)
		assert.NoError(t, err)
		payload = append(payload, valueBytes...)
	}

	// 构造完整报文
	p, err := packet.NewPacket(0x12345678, types.CmdUpload, payload)
	assert.NoError(t, err)

	// 编码报文
	encoded, err := packetCodec.EncodePacket(p)
	assert.NoError(t, err)

	// 解码并验证
	decoded, err := packetCodec.DecodePacket(encoded)
	assert.NoError(t, err)

	// 验证结果
	assert.Equal(t, p.Header.Length, decoded.Header.Length)
	assert.Equal(t, p.Data[:len(p.Data)], decoded.Data[:len(decoded.Data)])
	assert.Equal(t, p.CRC, decoded.CRC)
}
