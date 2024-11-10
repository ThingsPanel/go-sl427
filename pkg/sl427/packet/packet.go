// pkg/sl427/packet/packet.go
package packet

import (
	"github.com/ThingsPanel/go-sl427/pkg/sl427/types"
)

// Packet 表示一个完整的数据包,关注语义而不是字节格式
type Packet struct {
	Head        types.Header    // 帧头
	UserDataRaw []byte          // 数据域
	UserData    *types.UserData // 用户数据区
	CS          byte            // 校验码(CRC)
	EndFlag     byte            // 帧结束标识
	DataRaw     []byte          // 原始数据

}

// ParseUserData 解析用户数据区
// codec已经处理了帧格式,这里只需要处理用户数据区
func ParseUserData(frame *types.Frame) (*Packet, error) {
	// 解析用户数据区
	userData, err := types.NewUserData(frame.UserDataRaw)
	if err != nil {
		return nil, err
	}

	return &Packet{
		Head:        frame.Head,
		UserDataRaw: frame.UserDataRaw,
		UserData:    userData,
		CS:          frame.CS,
		EndFlag:     frame.EndFlag,
		DataRaw:     frame.Raw(),
	}, nil
}
