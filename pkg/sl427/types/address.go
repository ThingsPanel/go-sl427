// pkg/sl427/types/address.go

package types

import (
	"encoding/binary"
	"fmt"
)

// 地址域格式类型
const (
	AddrFormatType1 = 0x01 // 方式1:行政区划码+站点地址
	AddrFormatType2 = 0x00 // 方式2:特征码+站点编码
)

// 地址域长度定义
const (
	AdminCodeLen   = 3 // 行政区划码长度(方式1)
	StationAddrLen = 2 // 站点地址长度(方式1)
	FeatureCode    = 0 // 特征码(方式2)
)

// 站点地址范围(方式1)
const (
	MinStationAddr = 1     // 监测站最小地址
	MaxStationAddr = 60000 // 监测站最大地址
	MinRelayAddr   = 60001 // 中继站最小地址
	MaxRelayAddr   = 65534 // 中继站最大地址
	BroadcastAddr  = 65535 // 广播地址
	InvalidAddr    = 0     // 无效地址
)

// 地址域掩码(方式2)
// 每个字节被分成高4位和低4位来存储两个16进制数字。掩码(Mask)和位移(Shift)用于从字节中提取这些数字。
const (
	StationCodeMask  = 0x0F // 站点编码掩码
	StationCodeShift = 4    // 站点编码位移量
)

// Address 定义地址域接口
type Address interface {
	// Bytes 返回5字节的地址域二进制表示
	Bytes() []byte
	// Format 返回格式类型(1或2)
	Format() int
	// String 返回可读的字符串表示
	String() string
	// Validate 验证地址有效性
	Validate() error
	// 获站点地址或站点编码
	GetAddress() string
}

// AddressV1 方式1的地址实现(行政区划码 + 站点地址)
type AddressV1 struct {
	AdminCode []byte // 3字节BCD格式的行政区划码
	StationID uint16 // 2字节二进制格式的站点地址
}

// Bytes 实现Address接口
func (a *AddressV1) Bytes() []byte {
	buf := make([]byte, AddressLen)
	// 复制3字节行政区划码
	copy(buf[0:3], a.AdminCode)
	// 写入2字节站点地址
	binary.BigEndian.PutUint16(buf[3:], a.StationID)
	return buf
}

func (a *AddressV1) Format() int {
	return 1
}

func (a *AddressV1) String() string {
	return fmt.Sprintf("V1{AdminCode:%X,StationID:%d}", a.AdminCode, a.StationID)
}

func (a *AddressV1) Validate() error {
	// 1. 检查行政区划码长度
	if len(a.AdminCode) != 3 {
		return fmt.Errorf("行政区划码长度错误: %d", len(a.AdminCode))
	}

	// 2. 检查BCD码有效性
	for _, b := range a.AdminCode {
		if !BCD.IsValid(b) {
			return fmt.Errorf("无效的BCD码: %X", b)
		}
	}

	// 3. 检查站点地址范围
	switch {
	case a.StationID == InvalidAddr:
		return fmt.Errorf("无效的站点地址: 0")
	case a.StationID <= MaxStationAddr:
		// 监测站地址范围: 1-60000
		return nil
	case a.StationID <= MaxRelayAddr:
		// 中继站地址范围: 60001-65534
		return nil
	case a.StationID == BroadcastAddr:
		// 广播地址: 65535
		return nil
	default:
		return fmt.Errorf("站点地址超出范围: %d", a.StationID)
	}
}

func (a *AddressV1) GetAddress() string {
	return fmt.Sprintf("%s%04d", a.AdminCode, a.StationID)
}

// NewAddressV1 创建方式1的地址
func NewAddressV1(adminCode []byte, stationID uint16) (*AddressV1, error) {
	addr := &AddressV1{
		AdminCode: make([]byte, 3),
		StationID: stationID,
	}
	copy(addr.AdminCode, adminCode)

	if err := addr.Validate(); err != nil {
		return nil, err
	}
	return addr, nil
}

// AddressV2 方式2的地址实现(特征码 + 站点编码)
type AddressV2 struct {
	StationCode []byte // 4字节HEX格式的站点编码
}

// Bytes 实现Address接口
func (a *AddressV2) Bytes() []byte {
	buf := make([]byte, AddressLen)
	// 第1字节为特征码00H
	buf[0] = FeatureCode
	// 复制4字节站点编码
	copy(buf[1:], a.StationCode)
	return buf
}

func (a *AddressV2) Format() int {
	return 2
}

func (a *AddressV2) String() string {
	return fmt.Sprintf("V2{StationCode:%X}", a.StationCode)
}

func (a *AddressV2) Validate() error {
	// 1. 检查站点编码长度
	if len(a.StationCode) != 4 {
		return fmt.Errorf("站点编码长度错误: %d", len(a.StationCode))
	}

	// 2. 检查编码格式(每个半字节是否都有效)
	for i, b := range a.StationCode {
		high := b >> 4
		low := b & 0x0F
		if high > 0x0F || low > 0x0F {
			return fmt.Errorf("无效的HEX编码: byte[%d]=%X", i, b)
		}
	}

	return nil
}

func (a *AddressV2) GetAddress() string {
	// StationCode是4字节,每个字节分为高4位和低4位,总共是8位16进制数字
	// 例如: 0x80 0x00 0x00 0x01 应该解析为 "80000001"

	result := make([]byte, 8) // 8个16进制数字
	for i := 0; i < 4; i++ {
		// 处理每个字节的高4位和低4位
		high := (a.StationCode[i] >> 4) & 0x0F // 取高4位
		low := a.StationCode[i] & 0x0F         // 取低4位

		// 转换为ASCII字符
		result[i*2] = hexChar(high)
		result[i*2+1] = hexChar(low)
	}

	return string(result)
}

// hexChar 将4位二进制转换为16进制字符
func hexChar(n byte) byte {
	if n < 10 {
		return '0' + n // 0-9
	}
	return 'A' + (n - 10) // A-F
}

// NewAddressV2 创建方式2的地址
func NewAddressV2(stationCode []byte) (*AddressV2, error) {
	addr := &AddressV2{
		StationCode: make([]byte, 4),
	}
	copy(addr.StationCode, stationCode)

	if err := addr.Validate(); err != nil {
		return nil, err
	}
	return addr, nil
}

// ParseAddress 从字节流解析地址
func ParseAddress(data []byte) (Address, error) {
	if len(data) != AddressLen {
		return nil, fmt.Errorf("地址长度错误: %d", len(data))
	}

	// 根据第一个字节判断格式
	if data[0] == FeatureCode {
		// 方式2: 特征码为00H
		return NewAddressV2(data[1:])
	}

	// 方式1: 前3字节为行政区划码
	return NewAddressV1(
		data[0:3],
		binary.BigEndian.Uint16(data[3:]),
	)
}
