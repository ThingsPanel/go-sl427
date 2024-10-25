// pkg/sl427/metrics/metrics.go
package metrics

import (
	"sync/atomic"
	"time"
)

// Metrics 定义监控指标
type Metrics struct {
	PacketsReceived   uint64        // 接收的数据包数量
	PacketsSent       uint64        // 发送的数据包数量
	PacketsDropped    uint64        // 丢弃的数据包数量
	LastReceiveTime   atomic.Value  // 最后接收时间
	LastTransmitTime  atomic.Value  // 最后发送时间
	ProcessingLatency time.Duration // 处理延迟
}

// NewMetrics 创建新的监控指标实例
func NewMetrics() *Metrics {
	m := &Metrics{}
	m.LastReceiveTime.Store(time.Now())
	m.LastTransmitTime.Store(time.Now())
	return m
}

// RecordReceive 记录数据包接收
func (m *Metrics) RecordReceive() {
	atomic.AddUint64(&m.PacketsReceived, 1)
	m.LastReceiveTime.Store(time.Now())
}

// RecordSend 记录数据包发送
func (m *Metrics) RecordSend() {
	atomic.AddUint64(&m.PacketsSent, 1)
	m.LastTransmitTime.Store(time.Now())
}

// RecordDrop 记录数据包丢弃
func (m *Metrics) RecordDrop() {
	atomic.AddUint64(&m.PacketsDropped, 1)
}

// RecordLatency 记录处理延迟
func (m *Metrics) RecordLatency(start time.Time) {
	m.ProcessingLatency = time.Since(start)
}
