// pkg/sl427/types/measurement.go
package types

import (
	"encoding/json"
	"fmt"
)

// 注册函数
var parseUploadFuncMap = map[byte]func(byte, []byte) (json.RawMessage, error){
	DataTypeRain:       parseRain,
	DataTypeWaterLevel: parseWaterLevel,
}

// DeviceMode 确认帧的数据域,终端机工作模式
const (
	ModeCompatible = 0x00 // 兼容工作状态
	ModeUpload     = 0x01 // 自报工作状态
	ModeQuery      = 0x02 // 查询/应答工作状态
	ModeDebug      = 0x03 // 调试/维修状态
)

// DeviceStatus 设备状态(4字节)(AFN=81H)
type DeviceStatus struct {
	Alarm uint16 // 报警状态(2字节)
	State uint16 // 终端机状态(2字节)
}

// UploadFrame 自报数据帧
type UploadFrame struct {
	RawData []byte          // 原始数据
	Items   json.RawMessage // 数据项
	Status  DeviceStatus    // 状态信息
}

// ParseUploadData 解析自报数据的数据域D
// dataType 控制域C中的命令与类型码
// dataField 数据域D的原始字节流
// dataMap 数据项映射表:[命令与类型码]json的key
func ParseUploadData(dataType byte, dataField []byte) (*UploadFrame, error) {
	// 获取解析函数
	parseFunc, ok := parseUploadFuncMap[dataType]
	if !ok {
		return nil, fmt.Errorf("未找到解析函数，不支持的类型码: %d", dataType)
	}

	// 解析数据
	items, err := parseFunc(dataType, dataField)
	if err != nil {
		return nil, err
	}

	// 解析状态信息
	status := DeviceStatus{
		Alarm: uint16(dataField[0])<<8 | uint16(dataField[1]),
		State: uint16(dataField[2])<<8 | uint16(dataField[3]),
	}

	// 创建自报数据帧
	return &UploadFrame{
		RawData: dataField,
		Items:   items,
		Status:  status,
	}, nil
}

// ParseRain 解析雨量数据(3字节BCD码)
func parseRain(dataType byte, data []byte) (json.RawMessage, error) {
	if len(data) != 3 {
		return nil, fmt.Errorf("invalid rain data length: %d", len(data))
	}

	// 解析3字节BCD码为雨量值
	value := BCD.DecodeInt(data)

	// 转换为json格式
	return json.Marshal(map[string]interface{}{
		"YL": float64(value) / 10.0, // 保留一位小数
	})
}

// ParseWaterLevel 解析水位数据(每个水位4字节BCD码)
func parseWaterLevel(dataType byte, data []byte) (json.RawMessage, error) {
	if len(data) < 4 || len(data)%4 != 0 {
		return nil, fmt.Errorf("invalid water level data length: %d", len(data))
	}

	// 构造结果map
	result := make(map[string]interface{})

	// 计算水位个数
	count := len(data) / 4

	// 解析每个水位
	for i := 0; i < count; i++ {
		// 获取当前水位数据
		offset := i * 4
		levelData := data[offset : offset+4]

		// 解析符号位
		signByte := levelData[3] >> 4
		negative := signByte == 0x0F

		// 解析水位值
		value := float64(BCD.FromBCD(levelData[0]&0x0F))*0.001 + // 毫米位
			float64(BCD.FromBCD(levelData[0]>>4))*0.01 + // 厘米位
			float64(BCD.FromBCD(levelData[1]&0x0F))*0.1 + // 分米位
			float64(BCD.FromBCD(levelData[1]>>4)) + // 米个位
			float64(BCD.FromBCD(levelData[2]&0x0F))*10 + // 米十位
			float64(BCD.FromBCD(levelData[2]>>4))*100 + // 米百位
			float64(BCD.FromBCD(levelData[3]&0x0F))*1000 // 米千位

		if negative {
			value = -value
		}

		// 生成key(第一个用SW,后续用SW2,SW3...)
		key := "SW"
		if i > 0 {
			key = fmt.Sprintf("SW%d", i+1)
		}

		result[key] = value
	}

	return json.Marshal(result)
}
