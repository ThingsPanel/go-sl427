// pkg/sl427/types/basic.go
package types

import (
	"time"

	"github.com/ThingsPanel/go-sl427/pkg/sl427"
)

// 协议常量
const (
	StartFlag byte = 0x68 // 报文起始标识
	EndFlag   byte = 0x16 // 报文结束标识

	MaxDataLen   = 1024 // 最大数据长度
	MinDataLen   = 12   // 最小数据长度
	HeaderLen    = 10   // 报文头长度
	TimestampLen = 12   // 时间戳长度
)

// 命令类型定义
const (
	CmdQuery     byte = 0x01 // 查询命令
	CmdUpload    byte = 0x02 // 上传数据
	CmdHeartbeat byte = 0x03 // 心跳包
)

// 应答码定义
const (
	RespSuccess byte = 0x00 // 成功
	RespError   byte = 0x01 // 失败
)

// 设备状态定义
const (
	StatusNormal  byte = 0x00 // 正常
	StatusError   byte = 0x01 // 异常
	StatusOffline byte = 0x02 // 离线
)

// 数据类型定义
const (
	TypeInt8   byte = 0x01 // 8位整数
	TypeInt16  byte = 0x02 // 16位整数
	TypeInt32  byte = 0x03 // 32位整数
	TypeString byte = 0x04 // 字符串
	TypeTime   byte = 0x05 // 时间戳
)

// TimeStamp 时间戳类型(YYMMDDhhmmss)
type TimeStamp struct {
	time.Time
}

// DataValue 数据值
type DataValue struct {
	Type  byte   // 数据类型
	Value []byte // 原始数据
}

// NewTimeStamp 创建时间戳
func NewTimeStamp(t time.Time) TimeStamp {
	return TimeStamp{Time: t}
}

// Bytes 将时间转换为字节数组(YYMMDDhhmmss)
func (t TimeStamp) Bytes() []byte {
	return []byte(t.Format("060102150405"))
}

// ParseTimeStamp 解析时间戳
func ParseTimeStamp(data []byte) (TimeStamp, error) {
	if len(data) < TimestampLen {
		return TimeStamp{}, sl427.ErrInvalidLength
	}
	t, err := time.ParseInLocation("060102150405", string(data[:TimestampLen]), time.Local)
	if err != nil {
		return TimeStamp{}, sl427.ErrInvalidFormat
	}
	return TimeStamp{Time: t}, nil
}

// NewDataValue 创建数据值
func NewDataValue(dataType byte, value []byte) (*DataValue, error) {
	if len(value) > MaxDataLen {
		return nil, sl427.ErrInvalidLength
	}
	return &DataValue{
		Type:  dataType,
		Value: value,
	}, nil
}
