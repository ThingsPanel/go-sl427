// cmd/examples/basic/main.go
package main

import (
	"encoding/binary"
	"log"
	"time"

	"github.com/ThingsPanel/go-sl427/pkg/sl427/codec"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/packet"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
)

func main() {
	// 运行所有示例
	sendDataExample()
	receiveDataExample()
	handleHeartbeatExample()
}

// 发送数据示例
func sendDataExample() {
	log.Println("运行发送数据示例...")

	// 构造测量数据包
	var payload []byte

	// 1. 添加时间戳
	timestamp := types.NewTimeStamp(time.Now())
	payload = append(payload, timestamp.Bytes()...)

	// 2. 添加一个Int16类型的测量值
	payload = append(payload, types.TypeInt16) // 数据类型标识
	value := uint16(1234)                      // 测量值
	valueBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(valueBuf, value)
	payload = append(payload, valueBuf...)

	// 3. 创建数据包
	p, err := packet.NewPacket(0x12345678, types.CmdUpload, payload)
	if err != nil {
		log.Printf("创建数据包失败: %v", err)
		return
	}

	// 4. 编码数据包
	codec := codec.NewPacketCodec()
	encoded, err := codec.EncodePacket(p)
	if err != nil {
		log.Printf("编码失败: %v", err)
		return
	}

	log.Printf("数据包已编码: %X", encoded)

	// 5. 解码验证
	decoded, err := codec.DecodePacket(encoded)
	if err != nil {
		log.Printf("解码失败: %v", err)
		return
	}

	// 6. 解析数据内容
	data := decoded.Data
	if len(data) >= types.TimestampLen {
		ts, err := types.ParseTimeStamp(data[:types.TimestampLen])
		if err != nil {
			log.Printf("解析时间戳失败: %v", err)
			return
		}
		log.Printf("时间戳: %v", ts.Time)

		// 解析测量值
		if len(data) >= types.TimestampLen+3 { // timestampLen + type(1) + value(2)
			dataType := data[types.TimestampLen]
			if dataType == types.TypeInt16 {
				measureValue := binary.BigEndian.Uint16(data[types.TimestampLen+1:])
				log.Printf("测量值: %d", measureValue)
			}
		}
	}
}

// 接收数据示例
func receiveDataExample() {
	log.Println("运行接收数据示例...")

	// 创建一个模拟的数据包
	timestamp := types.NewTimeStamp(time.Now())
	mockPayload := append(timestamp.Bytes(), types.TypeInt16)
	mockValue := uint16(1234)
	valueBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(valueBuf, mockValue)
	mockPayload = append(mockPayload, valueBuf...)

	p, _ := packet.NewPacket(0x12345678, types.CmdUpload, mockPayload)
	codec := codec.NewPacketCodec()
	mockData, _ := codec.EncodePacket(p)

	// 解码数据包
	decoded, err := codec.DecodePacket(mockData)
	if err != nil {
		log.Printf("解码失败: %v", err)
		return
	}

	// 解析数据内容
	data := decoded.Data
	if len(data) >= types.TimestampLen {
		ts, _ := types.ParseTimeStamp(data[:types.TimestampLen])
		log.Printf("接收到数据 - 时间戳: %v", ts.Time)

		if len(data) >= types.TimestampLen+3 {
			dataType := data[types.TimestampLen]
			if dataType == types.TypeInt16 {
				value := binary.BigEndian.Uint16(data[types.TimestampLen+1:])
				log.Printf("接收到数据 - 测量值: %d", value)
			}
		}
	}
}

// 心跳包处理示例
func handleHeartbeatExample() {
	log.Println("运行心跳包示例...")

	// 创建心跳包，心跳包payload中只包含时间戳
	timestamp := types.NewTimeStamp(time.Now())
	p, _ := packet.NewPacket(0x12345678, types.CmdHeartbeat, timestamp.Bytes())

	// 编码发送
	codec := codec.NewPacketCodec()
	encoded, _ := codec.EncodePacket(p)

	log.Printf("发送心跳包: %X", encoded)

	// 模拟接收并处理
	received, err := codec.DecodePacket(encoded)
	if err != nil {
		log.Printf("解码心跳包失败: %v", err)
		return
	}

	if received.Header.Command == types.CmdHeartbeat {
		ts, _ := types.ParseTimeStamp(received.Data)
		log.Printf("收到心跳包 - 地址: %X, 时间: %v", received.Header.Address, ts.Time)
	}
}
