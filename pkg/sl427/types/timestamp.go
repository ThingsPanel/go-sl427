// pkg/sl427/types/timestamp.go
package types

import (
	"fmt"
	"time"
)

const TimestampLen = 7 // 6字节BCD时间 + 1字节超时

// Timestamp 时间标签结构(7字节)
type TimeLabel struct {
	Second  byte // 秒(BCD码)
	Minute  byte // 分(BCD码)
	Hour    byte // 时(BCD码)
	Day     byte // 日(BCD码)
	Month   byte // 月(BCD码)
	Year    byte // 年(BCD码,0-99)
	Timeout byte // 超时时间(分钟,BIN码)
}

// NewTimestamp 创建新的时间标签
func NewTimestamp(t time.Time) *TimeLabel {
	return &TimeLabel{
		Second:  BCD.ToBCD(byte(t.Second())),
		Minute:  BCD.ToBCD(byte(t.Minute())),
		Hour:    BCD.ToBCD(byte(t.Hour())),
		Day:     BCD.ToBCD(byte(t.Day())),
		Month:   BCD.ToBCD(byte(t.Month())),
		Year:    BCD.ToBCD(byte(t.Year() % 100)), // 只取年份后两位
		Timeout: 0,                               // 默认超时为0
	}
}

// Bytes 返回时间标签的字节表示
func (t *TimeLabel) Bytes() []byte {
	return []byte{
		t.Second,
		t.Minute,
		t.Hour,
		t.Day,
		t.Month,
		t.Year,
		t.Timeout,
	}
}

// ParseTimestamp 从字节数组解析时间标签
func ParseTimestamp(data []byte) (*TimeLabel, error) {
	if len(data) != TimestampLen {
		return nil, fmt.Errorf("invalid timestamp length: %d", len(data))
	}

	return &TimeLabel{
		Second:  data[0],
		Minute:  data[1],
		Hour:    data[2],
		Day:     data[3],
		Month:   data[4],
		Year:    data[5],
		Timeout: data[6],
	}, nil
}

func (t *TimeLabel) Seconds() int64 {
	// 将BCD码转换为实际数值
	second := BCD.FromBCD(t.Second)
	minute := BCD.FromBCD(t.Minute)
	hour := BCD.FromBCD(t.Hour)
	day := BCD.FromBCD(t.Day)
	month := BCD.FromBCD(t.Month)
	year := BCD.FromBCD(t.Year)

	// 构建完整时间
	// 注意：年份是两位数，需要转换为四位数
	// 这里假设年份范围是2000-2099
	fullYear := 2000 + int(year)

	// 使用time.Date构建时间对象
	timestamp := time.Date(
		fullYear,
		time.Month(month),
		int(day),
		int(hour),
		int(minute),
		int(second),
		0, // 纳秒部分为0
		time.Local,
	)

	return timestamp.Unix()
}

// 添加一个检查时间是否为零值的方法
func (t *TimeLabel) IsZero() bool {
	return t.Second == 0 && t.Minute == 0 && t.Hour == 0 &&
		t.Day == 0 && t.Month == 0 && t.Year == 0
}
