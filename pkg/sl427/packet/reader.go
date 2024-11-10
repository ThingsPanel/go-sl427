// pkg/sl427/packet/reader.go
package packet

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
)

// FrameReader 从io.Reader中读取SL427帧
type Reader struct {
	reader *bufio.Reader
}

// NewFrameReader 创建帧读取器
func NewReader(r io.Reader) *Reader {
	return &Reader{
		reader: bufio.NewReader(r),
	}
}

// ReadFrame 读取完整的SL427帧
func (r *Reader) ReadFrame() ([]byte, error) {
	var buf bytes.Buffer

	// 1. 查找起始标识
	startByte, err := r.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read start flag failed: %w", err)
	}

	// 寻找帧头
	if startByte != types.StartFlag {
		for {
			b, err := r.reader.ReadByte()
			if err != nil {
				return nil, err
			}
			if b == types.StartFlag {
				startByte = b
				break
			}
		}
	}
	buf.WriteByte(startByte)

	// 2. 读取长度字节
	length, err := r.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read length failed: %w", err)
	}
	buf.WriteByte(length)

	// 3. 读取第二个起始标识
	startByte2, err := r.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read second start flag failed: %w", err)
	}
	buf.WriteByte(startByte2)

	if startByte2 != types.StartFlag {
		return nil, fmt.Errorf("invalid second start flag")
	}

	// 4. 读取用户数据区和校验码
	remainingBytes := int(length) + 2 // 用户数据区 + CS + EndFlag
	data := make([]byte, remainingBytes)
	n, err := io.ReadFull(r.reader, data)
	if err != nil {
		return nil, fmt.Errorf("read remaining data failed: %w", err)
	}
	if n != remainingBytes {
		return nil, fmt.Errorf("incomplete frame data")
	}
	buf.Write(data)

	return buf.Bytes(), nil
}
