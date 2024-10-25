// pkg/sl427/types/data_item_test.go
package types

import (
	"testing"
)

func TestDataItemRegistry(t *testing.T) {
	// 测试注册和获取数据项
	registry := NewDataItemRegistry()

	def := DataItemDef{
		ID:          1001,
		Name:        "水位",
		Type:        TypeInt32,
		Unit:        "m",
		Scale:       -3,
		Description: "站点水位",
	}

	registry.Register(def)

	// 测试获取已注册的数据项
	got, ok := registry.Get(1001)
	if !ok {
		t.Error("未找到已注册的数据项")
		return
	}

	if got.ID != def.ID || got.Name != def.Name || got.Type != def.Type {
		t.Error("获取的数据项定义不匹配")
	}

	// 测试格式化值
	val := int32(12345) // 12.345m
	formatted := got.FormatValue(val)
	if formatted != "12.345m" {
		t.Errorf("格式化值错误, got %s, want 12.345m", formatted)
	}
}
