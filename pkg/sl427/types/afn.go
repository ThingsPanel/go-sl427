// pkg/sl427/types/afn.go

package types

import "fmt"

// AFN 功能码类型
type AFN byte

// 功能码定义 - 自报相关
const (
	AFNUpload    AFN = 0xC0 // 自报实时数据
	AFNAlarm     AFN = 0x81 // 自报告警数据
	AFNManualSet AFN = 0x82 // 人工置数
	AFNImageData AFN = 0x83 // 自报图片数据
	AFNVoltage   AFN = 0x84 // 自报电压数据
)

// IsValid 检查功能码是否有效
func (a AFN) IsValid() bool {
	switch a {
	case AFNUpload, AFNAlarm, AFNManualSet, AFNImageData, AFNVoltage:
		return true
	default:
		return false
	}
}

// String 返回功能码的字符串表示
func (a AFN) String() string {
	switch a {
	case AFNUpload:
		return "自报实时数据(0xC0)"
	case AFNAlarm:
		return "自报告警数据(0x81)"
	case AFNManualSet:
		return "人工置数(0x82)"
	case AFNImageData:
		return "自报图片数据(0x83)"
	case AFNVoltage:
		return "自报电压数据(0x84)"
	default:
		return fmt.Sprintf("未知功能码(0x%02X)", byte(a))
	}
}
