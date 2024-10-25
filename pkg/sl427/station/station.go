// pkg/sl427/station/station.go
package station

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ThingsPanel/go-sl427/pkg/sl427/codec"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/packet"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
)

// Station 表示一个监测站点
type Station struct {
	address   uint32
	conn      net.Conn
	codec     *codec.PacketCodec
	serialNum byte
	running   bool
	mu        sync.Mutex
	stopCh    chan struct{}
	logger    types.Logger
}

// Config 站点配置
type Config struct {
	Address  uint32
	Server   string
	Interval time.Duration
}

// NewStation 创建新的站点
func NewStation(config Config) *Station {
	return &Station{
		address: config.Address,
		codec:   codec.NewPacketCodec(),
		stopCh:  make(chan struct{}),
		logger:  types.DefaultLogger,
	}
}

// SetLogger 设置日志接口
func (s *Station) SetLogger(l types.Logger) {
	if l != nil {
		s.logger = l
	}
}

// Start 启动站点
func (s *Station) Start(config Config) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	conn, err := net.Dial("tcp", config.Server)
	if err != nil {
		return fmt.Errorf("连接服务器失败: %v", err)
	}
	s.conn = conn

	s.logger.Printf("站点[%X]已连接到服务器: %s", s.address, config.Server)

	go s.heartbeatLoop()
	go s.uploadLoop(config.Interval)

	return nil
}

// Stop 停止站点
func (s *Station) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopCh)
	if s.conn != nil {
		s.conn.Close()
	}

	s.logger.Printf("站点[%X]已停止", s.address)
}

// sendHeartbeat 发送心跳包
func (s *Station) sendHeartbeat() error {
	// 1. 构建时间戳
	ts := types.NewTimeStamp(time.Now())
	tsBytes := ts.Bytes()

	// 2. 构建心跳包
	p, err := packet.NewPacket(s.address, types.CmdHeartbeat, tsBytes)
	if err != nil {
		return fmt.Errorf("创建心跳包失败: %v", err)
	}

	// 3. 设置序列号
	p.Header.SerialNum = s.nextSerialNum()

	// 4. 获取完整的字节数据
	data := p.Bytes()

	// 5. 记录日志
	s.logger.Printf("站点[%X]发送心跳包: 长度=%d, 数据=%X",
		s.address, len(data), data)

	// 6. 发送数据
	_, err = s.conn.Write(data)
	if err != nil {
		return fmt.Errorf("发送心跳包失败: %v", err)
	}

	s.logger.Printf("站点[%X]发送心跳包: 序号=%d", s.address, p.Header.SerialNum)
	return nil
}

// heartbeatLoop 心跳维持
func (s *Station) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.sendHeartbeat(); err != nil {
				s.logger.Printf("站点[%X]发送心跳失败: %v", s.address, err)
			}
		}
	}
}

// uploadLoop 数据上报
func (s *Station) uploadLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.uploadData(); err != nil {
				s.logger.Printf("站点[%X]上报数据失败: %v", s.address, err)
			}
		}
	}
}

// uploadData 上报数据
func (s *Station) uploadData() error {
	// 采集数据
	data := s.collectData()

	// 构建数据域
	payload := s.buildPayload(data)

	// 创建数据包
	p, err := packet.NewPacket(s.address, types.CmdUpload, payload)
	if err != nil {
		return fmt.Errorf("创建数据包失败: %v", err)
	}

	// 设置序列号
	p.Header.SerialNum = s.nextSerialNum()

	// 获取完整数据包
	packetData := p.Bytes()

	// 调试日志 - 添加更详细的内容
	s.logger.Printf("站点[%X]准备发送数据包:\n"+
		"  长度=%d\n"+
		"  载荷长度=%d\n"+
		"  序号=%d\n"+
		"  数据=%X",
		s.address, len(packetData), len(payload),
		p.Header.SerialNum, packetData)

	// 发送数据
	_, err = s.conn.Write(packetData)
	if err != nil {
		return fmt.Errorf("发送数据包失败: %v", err)
	}

	return nil
}

// buildPayload 构建数据包载荷
func (s *Station) buildPayload(data MeasureData) []byte {
	// 预分配缓冲区 - 合理估算大小
	payload := make([]byte, 0, types.TimestampLen+1+len(data.Values)*10)

	// 1. 添加时间戳
	timestamp := types.NewTimeStamp(data.Timestamp)
	payload = append(payload, timestamp.Bytes()...)

	// 2. 添加数据项数量
	payload = append(payload, byte(len(data.Values)))

	// 3. 添加各个数据项
	for _, item := range data.Values {
		// 数据项ID(2字节)
		idBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(idBytes, item.ID)
		payload = append(payload, idBytes...)

		// 数据类型(1字节)
		payload = append(payload, item.Type)

		// 对于字符串类型，需要先写入长度
		if item.Type == types.TypeString {
			length := len(item.Value)
			if length > 255 {
				s.logger.Printf("警告:字符串数据过长,将被截断: ID=%d, len=%d",
					item.ID, length)
				length = 255
				item.Value = item.Value[:255]
			}
			payload = append(payload, byte(length))
		}

		// 数据值
		payload = append(payload, item.Value...)
	}

	return payload
}

// MeasureData 测量数据结构
type MeasureData struct {
	Timestamp time.Time
	Values    []DataValue
}

// DataValue 数据值结构
type DataValue struct {
	ID    uint16
	Type  byte
	Value []byte
}

// collectData 采集数据
func (s *Station) collectData() MeasureData {
	return MeasureData{
		Timestamp: time.Now(),
		Values: []DataValue{
			// 水位 - int32类型(单位:mm)
			{
				Type: types.TypeInt32,
				ID:   1001,
				Value: func() []byte {
					buf := make([]byte, 4)
					binary.BigEndian.PutUint32(buf, 12345) // 12.345m
					return buf
				}(),
			},
			// 流量 - int32类型(单位:L/s)
			{
				Type: types.TypeInt32,
				ID:   1002,
				Value: func() []byte {
					buf := make([]byte, 4)
					binary.BigEndian.PutUint32(buf, 5678) // 5.678m³/s
					return buf
				}(),
			},
			// 水质 - int16类型(PH值*100)
			{
				Type: types.TypeInt16,
				ID:   1003,
				Value: func() []byte {
					buf := make([]byte, 2)
					binary.BigEndian.PutUint16(buf, 723) // PH 7.23
					return buf
				}(),
			},
			// 水温 - int16类型(温度*100)
			{
				Type: types.TypeInt16,
				ID:   1004,
				Value: func() []byte {
					buf := make([]byte, 2)
					binary.BigEndian.PutUint16(buf, 2456) // 24.56℃
					return buf
				}(),
			},
			// 设备状态描述 - 字符串类型
			{
				Type:  types.TypeString,
				ID:    1005,
				Value: []byte("normal"),
			},
		},
	}
}

// nextSerialNum 生成下一个流水号
func (s *Station) nextSerialNum() byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serialNum++
	return s.serialNum
}
