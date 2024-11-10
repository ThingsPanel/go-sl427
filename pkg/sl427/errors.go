// pkg/sl427/errors.go
package sl427

import (
	"errors"
	"fmt"
)

// ErrorCode 定义错误码类型
type ErrorCode int

const (
	// 基础错误码 (1000-1099)
	ErrCodeInvalidData ErrorCode = 1000 + iota
	ErrCodeInvalidLength
	ErrCodeInvalidFormat
	ErrCodeInvalidValue
	ErrCodeInvalidType

	// 报文相关错误 (1100-1199)
	ErrCodeInvalidStartFlag ErrorCode = 1100 + iota
	ErrCodeInvalidEndFlag
	ErrCodePacketTooShort
	ErrCodePacketTooLong
	ErrCodeInvalidChecksum
	ErrCodeInvalidControl
	ErrCodeInvalidAddress
	ErrCodeInvalidAFN

	// 传输相关错误 (1200-1299)
	ErrCodeConnectionFailed ErrorCode = 1200 + iota
	ErrCodeTimeout
	ErrCodeDataTooLong
	ErrCodeReadFailed
	ErrCodeWriteFailed
	ErrCodeConnectionClosed

	// 协议相关错误 (1300-1399)
	ErrCodeUnsupportedVersion ErrorCode = 1300 + iota
	ErrCodeInvalidPassword
	ErrCodeInvalidTimeLabel
	ErrCodeResponseTimeout
	ErrCodeInvalidResponse
)

// Error 定义统一的错误类型
type Error struct {
	Code    ErrorCode // 错误码
	Message string    // 错误信息
	Cause   error     // 原始错误
}

// Error 实现 error 接口
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 支持 errors.Unwrap
func (e *Error) Unwrap() error {
	return e.Cause
}

// NewError 创建新的错误
func NewError(code ErrorCode, message string) error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// WrapError 包装已有错误
func WrapError(code ErrorCode, message string, err error) error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// 预定义的错误实例
var (
	// 基础错误
	ErrInvalidData   = NewError(ErrCodeInvalidData, "无效的数据")
	ErrInvalidLength = NewError(ErrCodeInvalidLength, "无效的数据长度")
	ErrInvalidFormat = NewError(ErrCodeInvalidFormat, "无效的数据格式")
	ErrInvalidValue  = NewError(ErrCodeInvalidValue, "无效的值")
	ErrInvalidType   = NewError(ErrCodeInvalidType, "无效的数据类型")

	// 报文错误
	ErrInvalidStartFlag = NewError(ErrCodeInvalidStartFlag, "无效的起始标识")
	ErrInvalidEndFlag   = NewError(ErrCodeInvalidEndFlag, "无效的结束标识")
	ErrPacketTooShort   = NewError(ErrCodePacketTooShort, "报文长度过短")
	ErrPacketTooLong    = NewError(ErrCodePacketTooLong, "报文长度过长")
	ErrInvalidChecksum  = NewError(ErrCodeInvalidChecksum, "无效的校验和")
	ErrInvalidControl   = NewError(ErrCodeInvalidControl, "无效的控制域")
	ErrInvalidAddress   = NewError(ErrCodeInvalidAddress, "无效的地址域")
	ErrInvalidAFN       = NewError(ErrCodeInvalidAFN, "无效的功能码")

	// 传输错误
	ErrConnectionFailed = NewError(ErrCodeConnectionFailed, "连接失败")
	ErrTimeout          = NewError(ErrCodeTimeout, "操作超时")
	ErrDataTooLong      = NewError(ErrCodeDataTooLong, "数据过长")
	ErrReadFailed       = NewError(ErrCodeReadFailed, "读取数据失败")
	ErrWriteFailed      = NewError(ErrCodeWriteFailed, "写入数据失败")
	ErrConnectionClosed = NewError(ErrCodeConnectionClosed, "连接已关闭")

	// 协议错误
	ErrUnsupportedVersion = NewError(ErrCodeUnsupportedVersion, "不支持的协议版本")
	ErrInvalidPassword    = NewError(ErrCodeInvalidPassword, "无效的密码")
	ErrInvalidTimeLabel   = NewError(ErrCodeInvalidTimeLabel, "无效的时间标签")
	ErrResponseTimeout    = NewError(ErrCodeResponseTimeout, "响应超时")
	ErrInvalidResponse    = NewError(ErrCodeInvalidResponse, "无效的响应")
)

// IsErrorCode 检查错误是否属于指定错误码
func IsErrorCode(err error, code ErrorCode) bool {
	var e *Error
	if err == nil {
		return false
	}
	if errors.As(err, &e) {
		return e.Code == code
	}
	return false
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) ErrorCode {
	var e *Error
	if err == nil {
		return 0
	}
	if errors.As(err, &e) {
		return e.Code
	}
	return 0
}

// IsTimeout 判断是否为超时错误
func IsTimeout(err error) bool {
	return IsErrorCode(err, ErrCodeTimeout) ||
		IsErrorCode(err, ErrCodeResponseTimeout)
}

// IsConnectionError 判断是否为连接相关错误
func IsConnectionError(err error) bool {
	return IsErrorCode(err, ErrCodeConnectionFailed) ||
		IsErrorCode(err, ErrCodeConnectionClosed)
}

// IsDataError 判断是否为数据相关错误
func IsDataError(err error) bool {
	return IsErrorCode(err, ErrCodeInvalidData) ||
		IsErrorCode(err, ErrCodeInvalidLength) ||
		IsErrorCode(err, ErrCodeInvalidFormat) ||
		IsErrorCode(err, ErrCodeInvalidValue) ||
		IsErrorCode(err, ErrCodeInvalidType)
}
