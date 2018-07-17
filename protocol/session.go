package protocol

// ISession
type ISession interface {
	GetId() string          // 获取 ID
	SetId(string)           // 设置 ID

	GetIdInt32() int32          // 获取 ID
	SetIdInt32(int32)           // 设置 ID

	HandleBytes([]byte)     // 接收数据
	Response(uint32,[]byte) // 回复数据
	Push([]byte)            // 推送数据
	Kick()                  // 踢下线
	IsTimeout() bool        // 是否心跳超时
	Close()                 // 关闭会话
	IsOk() bool             // 状态是否正常
	HandlePacket(packet Packet) int // 处理数据包
	send([]byte)            // 发送数据

	onDataPacket([]byte)    // 收到data包

	GetOwner() interface{} // 获取 owner
	SetOwner(interface{})  // 设置 owner
	updateHeartBeat()      // 更新心跳

	AddCloseEventListener(func(session ISession))
}
