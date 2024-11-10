// pkg/sl427/types/frame.go
package types

// 基本帧格式常量
const (
	// 帧标识符
	StartFlag byte = 0x68 // 帧起始标识(固定值68H)
	EndFlag   byte = 0x16 // 帧结束标识(固定值16H)

	// 长度限制
	MinFrameLen = 7   // 最小帧长度(帧头3 + 最小用户数据区1 + CS 1 + 结束符1)
	MaxFrameLen = 255 // 用户数据区最大长度(规约定义)

	// 固定长度字段
	AddressLen   = 5 // 地址域固定5字节
	TimeLabelLen = 7 // 时间标签固定7字节
)

// Frame 完整的帧结构定义
// 规约7.2.1节 帧格式框架表3
type Frame struct {
	Head        Header // 帧头
	UserDataRaw []byte // 用户数据区原始字节
	CS          byte   // 校验码(CRC)
	EndFlag     byte   // 帧结束标识
}

// FrameHeader 帧头定义(3字节)
// 规约7.2.2节 帧起始符和长度定义
type Header struct {
	StartFlag1 byte // 帧起始标识1(68H)
	Length     byte // 用户数据区长度L(1~255)
	StartFlag2 byte // 帧起始标识2(68H)
}

// 计算frame的长度
func (f *Frame) Len() int {
	return len(f.UserDataRaw) + 7
}

// 返回原始数据
func (f *Frame) Raw() []byte {
	return append(append([]byte{f.Head.StartFlag1, f.Head.Length, f.Head.StartFlag2}, f.UserDataRaw...), f.CS, f.EndFlag)
}
