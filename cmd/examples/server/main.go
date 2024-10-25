// cmd/examples/server/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ThingsPanel/go-sl427/pkg/sl427/codec"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/metrics"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/packet"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/protocol"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/transport"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
)

func init() {
	// 注册示例数据项定义
	types.DefaultRegistry.RegisterBatch([]types.DataItemDef{
		{
			ID:          1001,
			Name:        "水位",
			Type:        types.TypeInt32,
			Unit:        "m",
			Scale:       -3,
			Description: "站点水位",
		},
		{
			ID:          1002,
			Name:        "流量",
			Type:        types.TypeInt32,
			Unit:        "m³/s",
			Scale:       -3,
			Description: "流量",
		},
		{
			ID:          1003,
			Name:        "PH值",
			Type:        types.TypeInt16,
			Unit:        "",
			Scale:       -2,
			Description: "水质PH值",
		},
		{
			ID:          1004,
			Name:        "水温",
			Type:        types.TypeInt16,
			Unit:        "℃",
			Scale:       -2,
			Description: "水温",
		},
		{
			ID:          1005,
			Name:        "设备状态",
			Type:        types.TypeString,
			Unit:        "",
			Scale:       0,
			Description: "设备运行状态",
		},
	})
}

// 服务器配置
type Config struct {
	ListenAddr    string
	ReadTimeout   int
	WriteTimeout  int
	MaxConns      int
	MaxPacketSize int
}

// 服务器结构
type Server struct {
	config   Config
	listener net.Listener
	metrics  *metrics.Metrics
	protocol protocol.Protocol
	conns    sync.Map
	logger   types.Logger
}

// 包处理器
type packetHandler struct {
	conn     net.Conn
	protocol protocol.Protocol
	codec    *codec.PacketCodec
	metrics  *metrics.Metrics
	logger   types.Logger
}

// 修改 packetHandler 的 HandlePacket 方法
func (h *packetHandler) HandlePacket(p *packet.Packet) error {
	start := time.Now()
	defer h.metrics.RecordLatency(start)

	h.metrics.RecordReceive()

	// 根据命令类型处理
	switch p.Header.Command {
	case types.CmdHeartbeat:
		// 心跳包处理逻辑保持不变
		resp, err := h.protocol.BuildResponsePacket(p, true)
		if err != nil {
			h.metrics.RecordDrop()
			return fmt.Errorf("构建心跳响应失败: %v", err)
		}

		if err := h.sendResponse(resp); err != nil {
			h.metrics.RecordDrop()
			return fmt.Errorf("发送心跳响应失败: %v", err)
		}

		h.metrics.RecordSend()
		h.logger.Printf("收到心跳包并响应: 地址=%X, 序号=%d", p.Header.Address, p.Header.SerialNum)
		return nil

	case types.CmdUpload:
		// 解析上传数据
		data, err := h.protocol.ParseUploadData(p)
		if err != nil {
			h.metrics.RecordDrop()
			return fmt.Errorf("解析上传数据失败: %v", err)
		}

		// 构建并发送响应
		resp, err := h.protocol.BuildResponsePacket(p, true)
		if err != nil {
			h.metrics.RecordDrop()
			return fmt.Errorf("构建上传响应失败: %v", err)
		}

		if err := h.sendResponse(resp); err != nil {
			h.metrics.RecordDrop()
			return fmt.Errorf("发送上传响应失败: %v", err)
		}

		h.metrics.RecordSend()
		// 使用新的格式化函数输出详细信息
		h.logger.Printf("收到上传数据并响应: 地址=%X%s",
			p.Header.Address,
			formatUploadData(data))
		return nil

	default:
		h.metrics.RecordDrop()
		return fmt.Errorf("未知命令: %X", p.Header.Command)
	}
}

// sendResponse 发送响应包
func (h *packetHandler) sendResponse(resp *packet.Packet) error {
	encoded, err := h.codec.EncodePacket(resp)
	if err != nil {
		return fmt.Errorf("编码响应失败: %v", err)
	}

	_, err = h.conn.Write(encoded)
	if err != nil {
		return fmt.Errorf("发送响应失败: %v", err)
	}

	return nil
}

// 创建新服务器
func NewServer(config Config) *Server {
	return &Server{
		config:   config,
		metrics:  metrics.NewMetrics(),
		protocol: protocol.New(protocol.WithVersion("SL427-2021")),
		logger:   log.Default(),
	}
}

// 启动服务器
func (s *Server) Start(ctx context.Context) error {
	var err error
	s.listener, err = net.Listen("tcp", s.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("监听失败: %v", err)
	}

	s.logger.Printf("服务器启动在 %s", s.config.ListenAddr)

	go s.acceptLoop(ctx)

	return nil
}

// 接受连接循环
func (s *Server) acceptLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				s.logger.Printf("接受连接失败: %v", err)
				continue
			}

			// 检查连接数限制
			if s.getConnCount() >= s.config.MaxConns {
				s.logger.Printf("达到最大连接数限制(%d)", s.config.MaxConns)
				conn.Close()
				continue
			}

			// 创建处理器
			handler := transport.NewHandler(
				conn,
				&packetHandler{
					conn:     conn,
					protocol: s.protocol,
					codec:    codec.NewPacketCodec(),
					metrics:  s.metrics,
					logger:   s.logger,
				},
				transport.WithMaxPacketSize(s.config.MaxPacketSize),
				transport.WithTimeout(s.config.ReadTimeout, s.config.WriteTimeout),
				transport.WithLogger(s.logger),
			)

			// 保存连接
			s.conns.Store(conn.RemoteAddr().String(), handler)

			// 启动处理
			go func() {
				defer s.conns.Delete(conn.RemoteAddr().String())
				if err := handler.Handle(); err != nil {
					s.logger.Printf("连接处理错误 [%s]: %v", conn.RemoteAddr(), err)
				}
			}()
		}
	}
}

// 停止服务器
func (s *Server) Stop() error {
	// 关闭监听器
	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("关闭监听器失败: %v", err)
	}

	// 关闭所有连接
	s.conns.Range(func(key, value interface{}) bool {
		handler := value.(transport.Handler)
		if err := handler.Close(); err != nil {
			s.logger.Printf("关闭连接失败 [%s]: %v", handler.RemoteAddr(), err)
		}
		return true
	})

	return nil
}

// 获取当前连接数
func (s *Server) getConnCount() int {
	count := 0
	s.conns.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func main() {
	// 解析命令行参数
	var config Config
	flag.StringVar(&config.ListenAddr, "addr", ":8080", "监听地址")
	flag.IntVar(&config.ReadTimeout, "read-timeout", 30, "读取超时时间(秒)")
	flag.IntVar(&config.WriteTimeout, "write-timeout", 30, "写入超时时间(秒)")
	flag.IntVar(&config.MaxConns, "max-conns", 1000, "最大连接数")
	flag.IntVar(&config.MaxPacketSize, "max-packet-size", 1024, "最大包大小")
	flag.Parse()

	// 创建服务器
	server := NewServer(config)

	// 创建context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动服务器
	if err := server.Start(ctx); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待退出信号
	<-sigChan

	// 优雅关闭
	cancel()
	if err := server.Stop(); err != nil {
		log.Printf("关闭服务器失败: %v", err)
	}
}

// formatUploadData 格式化上传数据日志
func formatUploadData(data *protocol.UploadData) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n时间戳: %v", data.Timestamp))
	sb.WriteString(fmt.Sprintf("\n数据项数量: %d", len(data.Items)))
	sb.WriteString("\n数据项详情:")

	for i, item := range data.Items {
		// 查找数据项定义
		if def, ok := types.DefaultRegistry.Get(item.ID); ok {
			// 使用数据项定义中的名称和格式化方法
			sb.WriteString(fmt.Sprintf("\n  [%d] %s: %s",
				i+1,
				def.Name,
				def.FormatValue(item.Value)))
		} else {
			// 未注册的数据项使用默认格式
			sb.WriteString(fmt.Sprintf("\n  [%d] 数据项(%d): %v",
				i+1,
				item.ID,
				item.Value))
		}
	}
	return sb.String()
}
