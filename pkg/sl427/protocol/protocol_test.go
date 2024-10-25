// pkg/sl427/protocol/protocol_test.go
package protocol

import (
	"testing"

	"github.com/ThingsPanel/go-sl427/pkg/sl427/packet"
	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
)

func TestParseUploadData(t *testing.T) {
	// 构造上传数据包内容
	data := []byte{
		0x32, 0x31, 0x30, 0x35, 0x32, 0x35, // 时间戳
		0x31, 0x35, 0x32, 0x35, 0x30, 0x30,
		0x01,       // 1个数据项
		0x03, 0xE9, // ID=1001
		0x03,                   // 类型=Int32
		0x00, 0x00, 0x30, 0x39, // 值=12345
	}

	uploadData, err := ParseUploadData(data)
	if err != nil {
		t.Fatalf("解析数据失败: %v", err)
	}

	// 验证基本内容
	if len(uploadData.Items) != 1 {
		t.Errorf("数据项数量错误: 期望1, 实际%d", len(uploadData.Items))
	}

	item := uploadData.Items[0]
	if item.ID != 1001 {
		t.Errorf("数据项ID错误: 期望1001, 实际%d", item.ID)
	}
}

func TestBuildResponsePacket(t *testing.T) {
	proto := New()

	// 测试构建响应包
	reqPkt, _ := packet.NewPacket(0x01, types.CmdHeartbeat, nil)
	resPkt, err := proto.BuildResponsePacket(reqPkt, true)

	if err != nil {
		t.Fatalf("构建响应包失败: %v", err)
	}

	// 验证响应包
	if resPkt.Header.Command != types.CmdHeartbeat {
		t.Error("响应包命令码错误")
	}

	if len(resPkt.Data) != 1 || resPkt.Data[0] != types.RespSuccess {
		t.Error("响应状态错误")
	}
}
