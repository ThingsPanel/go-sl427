// pkg/sl427/transport/handler_test.go
package transport

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/ThingsPanel/go-sl427/pkg/sl427/packet"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
)

// mockConn 模拟一个简单的连接
type mockConn struct {
	readBuf *bytes.Buffer
	closed  bool
}

func newMockConn(data []byte) *mockConn {
	return &mockConn{
		readBuf: bytes.NewBuffer(data),
	}
}

func (c *mockConn) Read(p []byte) (n int, err error) {
	return c.readBuf.Read(p)
}

func (c *mockConn) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (c *mockConn) Close() error                       { c.closed = true; return nil }
func (c *mockConn) LocalAddr() net.Addr                { return nil }
func (c *mockConn) RemoteAddr() net.Addr               { return nil }
func (c *mockConn) SetDeadline(t time.Time) error      { return nil }
func (c *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// mockHandler 模拟一个简单的包处理器
type mockHandler struct {
	receivedPackets []*packet.Packet
}

func (h *mockHandler) HandlePacket(p *packet.Packet) error {
	h.receivedPackets = append(h.receivedPackets, p)
	return nil
}

func TestReadValidPacket(t *testing.T) {
	// 构造一个有效的心跳包
	pkt, err := packet.NewPacket(0x01, types.CmdHeartbeat, []byte{
		0x32, 0x31, 0x30, 0x35, 0x32, 0x35,
		0x31, 0x35, 0x32, 0x35, 0x30, 0x30,
	})
	if err != nil {
		t.Fatalf("构建测试包失败: %v", err)
	}

	data := pkt.Bytes()
	if len(data) == 0 {
		t.Fatal("生成的测试数据为空")
	}

	// 设置模拟连接和处理器
	mockHandler := &mockHandler{}
	conn := newMockConn(data)
	handler := NewHandler(conn, mockHandler)

	// 验证包是否被正确处理
	handler.Handle()

	if len(mockHandler.receivedPackets) != 1 {
		t.Error("未接收到数据包")
		return
	}

	// 验证接收到的包内容
	receivedPkt := mockHandler.receivedPackets[0]
	if receivedPkt.Header.Command != types.CmdHeartbeat {
		t.Errorf("命令码不匹配: 期望 %d, 实际 %d",
			types.CmdHeartbeat, receivedPkt.Header.Command)
	}
}

func TestReadInvalidPacket(t *testing.T) {
	// 构造一个无效包(非0x68起始)
	invalidData := []byte{0x00, 0x01, 0x02}

	conn := newMockConn(invalidData)
	handler := NewHandler(conn, &mockHandler{})

	// 无效包应该返回错误
	err := handler.Handle()
	if err == nil {
		t.Error("处理无效包应该返回错误")
	}
}

func TestHandlerClose(t *testing.T) {
	conn := newMockConn(nil)
	handler := NewHandler(conn, &mockHandler{})

	err := handler.Close()
	if err != nil {
		t.Errorf("关闭处理器失败: %v", err)
	}

	if !conn.closed {
		t.Error("连接未被正确关闭")
	}
}
