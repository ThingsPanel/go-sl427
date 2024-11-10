// pkg/sl427/types/bcd.go

package types

// BCDCodec BCD编解码器
type BCDCodec struct{}

// BCD 全局BCD编解码器实例
var BCD = BCDCodec{}

// ToBCD 将数字转换为BCD编码
func (c BCDCodec) ToBCD(n byte) byte {
	// 十位和个位分别转BCD
	return (n/10)<<4 | (n % 10)
}

// FromBCD 将BCD编码转换为数字
func (c BCDCodec) FromBCD(b byte) byte {
	// 分别获取十位和个位,然后计算实际值
	return ((b>>4)&0x0F)*10 + (b & 0x0F)
}

// IsValid 检查是否有效的BCD码
func (c BCDCodec) IsValid(b byte) bool {
	// 检查高4位和低4位是否都在0-9范围内
	high := (b >> 4) & 0x0F
	low := b & 0x0F
	return high <= 9 && low <= 9
}

// Encode 将byte数组转换为BCD编码
func (c BCDCodec) Encode(data []byte) []byte {
	// 计算需要的BCD字节数
	bcdLen := (len(data) + 1) / 2
	bcd := make([]byte, bcdLen)

	for i := 0; i < len(data); i += 2 {
		// 高半字节
		high := (data[i] - '0') << 4
		if i+1 < len(data) {
			// 低半字节
			low := data[i+1] - '0'
			bcd[i/2] = high | low
		} else {
			// 最后一个字节填0
			bcd[i/2] = high
		}
	}
	return bcd
}

// Decode 将BCD编码转换回byte数组
func (c BCDCodec) Decode(bcd []byte) []byte {
	data := make([]byte, len(bcd)*2)

	for i := 0; i < len(bcd); i++ {
		// 解析高4位和低4位
		high := (bcd[i] >> 4) & 0x0F
		low := bcd[i] & 0x0F

		// 转换为ASCII字符
		data[i*2] = high + '0'
		data[i*2+1] = low + '0'
	}
	return data
}

// EncodeInt 将整数编码为BCD
func (c BCDCodec) EncodeInt(n uint32, bytes int) []byte {
	bcd := make([]byte, bytes)
	for i := bytes - 1; i >= 0; i-- {
		bcd[i] = c.ToBCD(byte(n % 100))
		n /= 100
	}
	return bcd
}

// DecodeInt 将BCD解码为整数
func (c BCDCodec) DecodeInt(bcd []byte) uint32 {
	var n uint32
	for _, b := range bcd {
		n = n*100 + uint32(c.FromBCD(b))
	}
	return n
}
