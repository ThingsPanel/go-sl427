// pkg/sl427/packet/reader.go
package packet

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/ThingsPanel/go-sl427/pkg/sl427/codec"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
)

// FrameReader 从io.Reader中读取SL427帧
type Reader struct {
	reader *bufio.Reader
	logger types.Logger
}

// NewFrameReader 创建帧读取器
func NewReader(r io.Reader, logger types.Logger) *Reader {
	return &Reader{
		reader: bufio.NewReader(r),
		logger: logger,
	}
}

func (r *Reader) ReadFrame() (*types.Frame, error) {
	var buf bytes.Buffer

	// 1. 查找起始标识
	startByte, err := r.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("读取起始标识失败: %w", err)
	}

	// 寻找帧头
	if startByte != types.StartFlag {
		for {
			b, err := r.reader.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("寻找起始标识时出错: %w", err)
			}
			if b == types.StartFlag {
				startByte = b
				break
			}
			// 记录跳过的无效字节
			r.logger.Printf("跳过无效字节: 0x%02X(期望为0x68)", b)
		}
	}
	buf.WriteByte(startByte)

	// 2. 读取长度字节
	length, err := r.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("读取长度字节失败: %w", err)
	}
	// 验证长度的合法性
	if length == 0 || length > types.MaxFrameLen {
		return nil, fmt.Errorf("无效的长度值: %d(应该在1-%d之间)", length, types.MaxFrameLen)
	}
	buf.WriteByte(length)

	// 3. 读取第二个起始标识
	startByte2, err := r.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("读取第二个起始标识失败: %w", err)
	}
	buf.WriteByte(startByte2)

	if startByte2 != types.StartFlag {
		return nil, fmt.Errorf("第二个起始标识错误: 0x%02X(期望值为0x68)", startByte2)
	}

	// 4. 读取用户数据区和校验码
	remainingBytes := int(length) + 2 // 用户数据区 + CS + EndFlag
	data := make([]byte, remainingBytes)
	n, err := io.ReadFull(r.reader, data)
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("数据不完整: 期望%d字节,实际读取%d字节", remainingBytes, n)
		}
		return nil, fmt.Errorf("读取剩余数据失败: %w", err)
	}

	// 检查结束标识
	if data[len(data)-1] != types.EndFlag {
		return nil, fmt.Errorf("结束标识错误: 0x%02X(期望值为0x16)", data[len(data)-1])
	}

	buf.Write(data)

	// 输出完整的数据包内容(用于调试)
	rawData := buf.Bytes()
	r.logger.Printf("读取到数据包: % X", rawData)

	codec := codec.NewPacketCodec()
	frame, err := codec.DecodePacket(rawData)
	if err != nil {
		return nil, fmt.Errorf("解码数据包失败[原始数据:% X]: %w", rawData, err)
	}

	return frame, nil
}
