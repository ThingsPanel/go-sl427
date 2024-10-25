// pkg/sl427/doc.go
/*
sl427包实现水资源监测的sl427 -2021协议。
它为监测站和数据收集服务器提供功能。

Example usage:
    station := sl427.NewStation(sl427.Config{
        Address: 0x01,
        Server: "localhost:8080",
        Interval: time.Second * 30,
    })
    if err := station.Start(); err != nil {
        log.Fatal(err)
    }
    defer station.Stop()
*/
package sl427
