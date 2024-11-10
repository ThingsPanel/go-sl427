// pkg/sl427/types/control.go

package types

import "fmt"

// 控制域定义
const (
	DirBit   = 0x80 // 传输方向位(D7) 1:上行（由终端机发出） 0:下行（由中心站发出）
	DivBit   = 0x40 // 拆分标志位(D6) 1:拆分 0:不拆分
	FcbMask  = 0x30 // 帧计数位(D5~D4) 00:帧计数0 01:帧计数1 10:帧计数2 11:帧计数3
	CodeMask = 0x0F // 命令与类型码掩码(D3~D0)
)

// 上行帧命令与类型码(DIR=1, D3~D0)，这里只定义了上行的命令与类型码
const (
	// 帧类型：确认
	CmdUpConfirm = 0x00 // 认可
	// 帧类型：自报帧
	DataTypeRain       = 0x01 // 雨量参数
	DataTypeWaterLevel = 0x02 // 水位参数
	DataTypeFlow       = 0x03 // 流量(水量)参数
	DataTypeSpeed      = 0x04 // 流速参数
	DataTypeGate       = 0x05 // 闸位参数
	DataTypePower      = 0x06 // 功率参数
	DataTypeWeather    = 0x07 // 气象参数
	DataTypeElectric   = 0x08 // 电量参数
	DataTypeTemp       = 0x09 // 水温参数
	DataTypeQuality    = 0x0A // 水质参数
	DataTypeSoil       = 0x0B // 土壤含水率参数
	DataTypeEvapor     = 0x0C // 蒸发量参数
	DataTypeAlarm      = 0x0D // 报警状态参数
	DataTypeRainStat   = 0x0E // 统计雨量
	DataTypePressure   = 0x0F // 水压参数
)

// Control 控制域结构体
type Control struct {
	value byte  // 第一个字节
	divs  *byte // 拆分帧计数(可选的第二个字节)
}

// NewControl 创建新的控制域
func NewControl(value byte) *Control {
	return &Control{value: value}
}

// SetDIV 设置拆分标志和计数
func (c *Control) SetDIV(count byte) {
	c.value |= 0x40 // 设置D6位
	c.divs = &count
}

// IsDIV 判断是否为拆分帧
func (c *Control) IsDIV() bool {
	return (c.value & 0x40) != 0
}

// DIR 获取传输方向(true表示上行,false表示下行)
func (c *Control) DIR() bool {
	return (c.value & 0x80) != 0
}

// SetDIR 设置传输方向
func (c *Control) SetDIR(up bool) {
	if up {
		c.value |= 0x80 // 设置D7位为1(上行)
	} else {
		c.value &^= 0x80 // 清除D7位(下行)
	}
}

// FCB 获取帧计数位(D5~D4)
func (c *Control) FCB() byte {
	return (c.value >> 4) & 0x03
}

// SetFCB 设置帧计数位
func (c *Control) SetFCB(fcb byte) {
	c.value = (c.value & 0xCF) | ((fcb & 0x03) << 4)
}

// Code 获取命令与类型码(D3~D0)
func (c *Control) Code() byte {
	return c.value & 0x0F
}

// SetCode 设置命令与类型码
func (c *Control) SetCode(code byte) {
	c.value = (c.value & 0xF0) | (code & 0x0F)
}

// Bytes 返回控制域的字节表示
func (c *Control) Bytes() []byte {
	if c.divs != nil {
		return []byte{c.value, *c.divs}
	}
	return []byte{c.value}
}

// Length 返回控制域长度(1或2字节)
func (c *Control) Length() int {
	if c.divs != nil {
		return 2
	}
	return 1
}

// IsUp 判断是否为上行
func (c *Control) IsUp() bool {
	return (c.value & DirBit) != 0
}

// GetType 获取数据类型
func (c *Control) GetType() byte {
	return c.value & CodeMask
}

// String 友好的字符串表示
func (c *Control) String() string {
	dir := "下行"
	if c.IsUp() {
		dir = "上行"
	}
	return fmt.Sprintf("%s类型:%d", dir, c.GetType())
}
