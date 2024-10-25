// pkg/sl427/types/data_item.go
package types

import (
	"fmt"
	"math"
	"strconv"
)

// DataItemDef 数据项定义
type DataItemDef struct {
	ID          uint16 // 数据项ID
	Name        string // 数据项名称
	Type        byte   // 数据类型
	Unit        string // 单位
	Scale       int    // 缩放因子(10的幂次), 如 -3 表示除以1000
	Description string // 描述
}

// DataItemRegistry 数据项注册表
type DataItemRegistry struct {
	items map[uint16]DataItemDef
}

// NewDataItemRegistry 创建数据项注册表
func NewDataItemRegistry() *DataItemRegistry {
	return &DataItemRegistry{
		items: make(map[uint16]DataItemDef),
	}
}

// Register 注册数据项定义
func (r *DataItemRegistry) Register(def DataItemDef) {
	r.items[def.ID] = def
}

// RegisterBatch 批量注册数据项定义
func (r *DataItemRegistry) RegisterBatch(defs []DataItemDef) {
	for _, def := range defs {
		r.Register(def)
	}
}

// Get 获取数据项定义
func (r *DataItemRegistry) Get(id uint16) (DataItemDef, bool) {
	def, ok := r.items[id]
	return def, ok
}

// FormatValue 根据数据项定义格式化值
func (def DataItemDef) FormatValue(value interface{}) string {
	scale := float64(1)
	if def.Scale != 0 {
		scale = math.Pow10(def.Scale)
	}

	switch def.Type {
	case TypeInt8:
		if v, ok := value.(int8); ok {
			return fmt.Sprintf("%."+strconv.Itoa(-def.Scale)+"f%s", float64(v)*scale, def.Unit)
		}
	case TypeInt16:
		if v, ok := value.(int16); ok {
			return fmt.Sprintf("%."+strconv.Itoa(-def.Scale)+"f%s", float64(v)*scale, def.Unit)
		}
	case TypeInt32:
		if v, ok := value.(int32); ok {
			return fmt.Sprintf("%."+strconv.Itoa(-def.Scale)+"f%s", float64(v)*scale, def.Unit)
		}
	case TypeString:
		if v, ok := value.(string); ok {
			return v
		}
	}
	return fmt.Sprintf("%v", value)
}

// DefaultRegistry 默认注册表实例
var DefaultRegistry = NewDataItemRegistry()
