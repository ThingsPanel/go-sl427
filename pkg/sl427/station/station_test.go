// pkg/sl427/station/station_test.go
package station

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
)

// mockConn 模拟连接
type mockConn struct {
	writeBuf bytes.Buffer
	closed   bool
}

func (m *mockConn) Read(b []byte) (n int, err error)   { return len(b), nil }
func (m *mockConn) Write(b []byte) (n int, err error)  { return m.writeBuf.Write(b) }
func (m *mockConn) Close() error                       { m.closed = true; return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestStationBasic(t *testing.T) {
	// 创建站点配置
	config := Config{
		Address:  0x01, // 站点地址
		Server:   "localhost:8080",
		Interval: time.Second,
	}

	// 创建站点
	station := NewStation(config)
	if station == nil {
		t.Fatal("创建站点失败")
	}

	// 启动和停止测试
	err := station.Start(config)
	if err != nil {
		t.Errorf("启动站点失败: %v", err)
	}

	station.Stop()
}

func TestStationHeartbeat(t *testing.T) {
	// 创建站点实例
	station := &Station{
		address: 0x01,
		conn:    &mockConn{},
		stopCh:  make(chan struct{}),
		logger:  types.DefaultLogger,
	}

	// 测试发送心跳
	err := station.sendHeartbeat()
	if err != nil {
		t.Errorf("发送心跳失败: %v", err)
	}
}
