// pkg/sl427/types/user_data.go
package types

import (
	"fmt"
	"strings"
)

// UserData 用户数据区定义(规约7.2.3节)
type UserData struct {
	Control   Control    // 控制域C(1或2字节)
	Address   Address    // 地址域A(5字节)
	AFN       AFN        // 功能码(1字节)
	UserAFN   *byte      // 用户功能码(1字节,可选)
	DataField []byte     // 数据域D的原始字节流
	PW        *byte      // 密码PW(2字节,可选)
	Tp        *TimeLabel // 时间标签Tp(7字节,可选)
}

// NewUserData 从字节流解析用户数据区
func NewUserData(data []byte) (*UserData, error) {
	if len(data) < 7 { // 最小长度:控制域(1)+地址域(5)+AFN(1)
		return nil, fmt.Errorf("数据长度不足: %d", len(data))
	}

	userData := &UserData{}
	offset := 0

	// 1. 解析控制域
	ctrl := NewControl(data[offset])
	if ctrl.IsDIV() {
		if len(data) < 8 { // 需要额外1字节
			return nil, fmt.Errorf("拆分帧数据长度不足")
		}
		ctrl.SetDIV(data[offset+1])
		offset += 2
	} else {
		offset++
	}
	userData.Control = *ctrl

	// 2. 解析地址域
	addr, err := ParseAddress(data[offset : offset+5])
	if err != nil {
		return nil, fmt.Errorf("解析地址域失败: %v", err)
	}
	userData.Address = addr
	offset += 5

	// 3. 解析功能码
	userData.AFN = AFN(data[offset])
	offset++

	// 4. 处理用户自定义功能码
	if userData.AFN == 0xFF {
		if offset >= len(data) {
			return nil, fmt.Errorf("解析用户功能码失败: 数据不足")
		}
		userAFN := data[offset]
		userData.UserAFN = &userAFN
		offset++
	}

	restData := data[offset:]

	// 5. 尝试解析时间标签(如果有)
	if len(restData) >= 7 {
		timeData := restData[len(restData)-7:]
		// 验证是否为有效的时间标签
		if isValidTimeLabel(timeData) {
			timestamp, err := ParseTimestamp(timeData)
			if err == nil {
				userData.Tp = timestamp
				restData = restData[:len(restData)-7]
			}
		}
	}

	// 6. 处理密码(如果存在)
	if !ctrl.DIR() && len(restData) >= 2 { // 下行报文可能包含密码
		pw := restData[len(restData)-2]
		userData.PW = &pw
		restData = restData[:len(restData)-2]
	}

	// 7. 保存剩余数据为数据域
	userData.DataField = restData

	return userData, nil
}

// isValidTimeLabel 简单验证是否为有效的时间标签
func isValidTimeLabel(data []byte) bool {
	if len(data) != 7 {
		return false
	}
	// 验证时分秒等是否在合理范围
	// 秒
	if !BCD.IsValid(data[0]) || BCD.FromBCD(data[0]) > 59 {
		return false
	}
	// 分
	if !BCD.IsValid(data[1]) || BCD.FromBCD(data[1]) > 59 {
		return false
	}
	// 时
	if !BCD.IsValid(data[2]) || BCD.FromBCD(data[2]) > 23 {
		return false
	}
	// 日
	if !BCD.IsValid(data[3]) || BCD.FromBCD(data[3]) > 31 || BCD.FromBCD(data[3]) == 0 {
		return false
	}
	// 月
	if !BCD.IsValid(data[4]) || BCD.FromBCD(data[4]) > 12 || BCD.FromBCD(data[4]) == 0 {
		return false
	}
	// 年和延时值不做特殊验证
	return true
}

// Bytes 将用户数据区编码为字节流
func (u *UserData) Bytes() []byte {
	// 计算总长度
	length := u.Control.Length() + 5 + 1 // 控制域 + 地址域 + AFN
	if u.UserAFN != nil {
		length++
	}
	if u.PW != nil {
		length += 2
	}
	length += len(u.DataField)
	if u.Tp != nil {
		length += 7
	}

	// 分配缓冲区
	buf := make([]byte, 0, length)

	// 1. 写入控制域
	buf = append(buf, u.Control.Bytes()...)

	// 2. 写入地址域
	buf = append(buf, u.Address.Bytes()...)

	// 3. 写入功能码
	buf = append(buf, byte(u.AFN))

	// 4. 写入用户功能码(如果存在)
	if u.UserAFN != nil {
		buf = append(buf, *u.UserAFN)
	}

	// 5. 写入数据域
	buf = append(buf, u.DataField...)

	// 6. 写入密码(如果存在)
	if u.PW != nil {
		buf = append(buf, *u.PW)
	}

	// 7. 写入时间标签(如果存在)
	if u.Tp != nil {
		buf = append(buf, u.Tp.Bytes()...)
	}

	return buf
}

// Validate 验证用户数据区的有效性
func (u *UserData) Validate() error {
	// 1. 验证地址
	if err := u.Address.Validate(); err != nil {
		return fmt.Errorf("无效的地址域: %v", err)
	}

	// 2. 验证功能码
	if !u.AFN.IsValid() {
		return fmt.Errorf("无效的功能码: %02X", u.AFN)
	}

	// 3. 验证用户功能码
	if u.AFN == 0xFF && u.UserAFN == nil {
		return fmt.Errorf("缺少用户功能码")
	}

	// 4. 验证密码(下行报文)
	if !u.Control.DIR() && u.PW == nil {
		return fmt.Errorf("下行报文缺少密码")
	}

	return nil
}

// String 返回用户数据区的可读字符串表示
func (u *UserData) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Control: %s\n", u.Control.String()))
	sb.WriteString(fmt.Sprintf("Address: %s\n", u.Address.String()))
	sb.WriteString(fmt.Sprintf("AFN: %s\n", u.AFN.String()))
	if u.UserAFN != nil {
		sb.WriteString(fmt.Sprintf("UserAFN: %02X\n", *u.UserAFN))
	}
	sb.WriteString(fmt.Sprintf("DataField: %X\n", u.DataField))
	if u.PW != nil {
		sb.WriteString(fmt.Sprintf("PW: %02X\n", *u.PW))
	}
	sb.WriteString(fmt.Sprintf("TimeLabel: %+v", u.Tp))
	return sb.String()
}
