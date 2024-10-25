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

	// 传输相关错误 (1200-1299)
	ErrCodeConnectionFailed ErrorCode = 1200 + iota
	ErrCodeTimeout
	ErrCodeDataTooLong
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
	ErrInvalidData   = NewError(ErrCodeInvalidData, "invalid data")
	ErrInvalidLength = NewError(ErrCodeInvalidLength, "invalid data length")
	ErrInvalidFormat = NewError(ErrCodeInvalidFormat, "invalid data format")
	ErrInvalidValue  = NewError(ErrCodeInvalidValue, "invalid value")
	ErrInvalidType   = NewError(ErrCodeInvalidType, "invalid data type")

	// 报文错误
	ErrInvalidStartFlag = NewError(ErrCodeInvalidStartFlag, "invalid start flag")
	ErrInvalidEndFlag   = NewError(ErrCodeInvalidEndFlag, "invalid end flag")
	ErrPacketTooShort   = NewError(ErrCodePacketTooShort, "packet too short")
	ErrPacketTooLong    = NewError(ErrCodePacketTooLong, "packet too long")
	ErrInvalidChecksum  = NewError(ErrCodeInvalidChecksum, "invalid checksum")

	// 传输错误
	ErrConnectionFailed = NewError(ErrCodeConnectionFailed, "connection failed")
	ErrTimeout          = NewError(ErrCodeTimeout, "operation timeout")
	ErrDataTooLong      = NewError(ErrCodeDataTooLong, "data too long")
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
