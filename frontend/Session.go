package frontend

import (
	"net"
	"time"
	"log"
	"fmt"
	"github.com/netc0/gate/common"
)

type Session struct {
	common.ISession
	holder interface{}
	id     string
	isOk   bool
	reader common.PacketReader
	time   time.Time // 心跳
	OnDataPacket func(common.ISession, uint32, uint32, []byte)

	owner interface{}

	closeEventListeners []func(common.ISession)
}

type TCPSession struct {
	Session
	conn net.Conn
}

type UDPSession struct {
	Session
	remote *net.UDPAddr
	conn *net.UDPConn
}

// 获取 ID
func (this *Session)GetId() string { return this.id }
// 设置 ID
func (this *Session) SetId(id string) { this.id = id }
// 接收数据
func (this *Session)HandleBytes(data[]byte){
	this.time = time.Now()
	var pkg = this.reader.ParsePacket(data)
	for {
		if pkg == nil {
			break
		}
		if this.HandlePacket(*pkg) != 0 { // pkg error, disconnect now
			log.Println("need close")
			this.Close()
			break
		}
		pkg = this.reader.ParsePacket(nil)
	}
}
// 回复数据
func (this *Session)Response(requestId uint32, r[]byte){
	var data = common.PacketResponseToBinary(common.PacketType_DATA, requestId, r)
	this.send(data) // 必须回应SYN
}
// 推送数据
func (this *Session)Push(data []byte){
	switch t := this.holder.(type){
	default:
		log.Println("Unknow session", this.holder)
	case TCPSession:
		t.send(data)
		break
	case UDPSession:
		t.send(data)
	}
}
// 踢下线
func (this *Session)Kick(){}
// 是否心跳超时
func (this *Session)IsTimeout() bool{ return time.Now().Second() - this.time.Second() > 5}
// 关闭会话
func (this *Session)Close(){
	switch t := this.holder.(type){
	default:
		this.isOk = false
		fmt.Println("know type: %v", t)
	case TCPSession:
		log.Println("close tcp conn")
		t.conn.Close()
	case UDPSession:
		log.Println("close udp conn")

		break
	}

	for _, callback := range this.closeEventListeners {
		callback(this)
	}

	RemoveSession(this)
}
// 状态是否正常
func (this *Session)IsOk() bool{ return false }
// 处理数据包
func (this *Session)HandlePacket(packet common.Packet) int {
	this.time = time.Now()
	if packet.Type == common.PacketType_SYN { // 收到 SYN
		var data = common.PacketToBinary(common.PacketType_ACK, nil)
		this.send(data) // 必须回应SYN
		return 0
	} else if packet.Type == common.PacketType_ACK { // 收到 ACK
		return 0
	} else if packet.Type == common.PacketType_HEARTBEAT { // 纯心跳包 一般不需要
		return 0
	} else if packet.Type == common.PacketType_DATA { // on data
		this.onDataPacket(packet.Body)
		return 0
	} else if packet.Type == common.PacketType_KICK { // on kick

	}
	log.Println("packet type not support")
	return -1
}
// 发送数据
func (this* Session) send(data[]byte) {
	switch t := this.holder.(type){
	default:
		log.Println("Unknow session", this.holder)
	case TCPSession:
		t.send(data)
		break
	case UDPSession:
		t.send(data)
	}
}
// 收到数据包
func (this* Session) onDataPacket(data []byte) {
	// [requestId] [routeId] [data]
	// 1. 解析出requestId
	var requestId uint32
	var routeId uint32

	requestId = uint32(
			uint32(data[0]) << 24 |
			uint32(data[1]) << 16 |
			uint32(data[2]) << 8 |
			uint32(data[3]));
	// 2. 解析出 routeId
	routeId = uint32(
			uint32(data[4]) << 24 |
			uint32(data[5]) << 16 |
			uint32(data[6]) << 8 |
			uint32(data[7]));
	data = data[8:]

	if this.OnDataPacket != nil {
		this.OnDataPacket(this, requestId, routeId, data)
	}
}
// 关闭会话的回调
func (this*Session) AddCloseEventListener(callback func(session common.ISession)) {
	this.closeEventListeners = append(this.closeEventListeners, callback)
}
// 发送 TCP 消息
func (this* TCPSession) send(data[]byte) {
	b, err := this.conn.Write(data)
	if err != nil {
		log.Println("tcp write: ", b, err)
	}
}
// 发送 UDP 消息
func (this* UDPSession) send(data[]byte) {
	b, err := this.conn.WriteToUDP(data, this.remote)
	if err != nil {
		log.Println("udp write", b, err)
	}
}
// 获取 owner
func (this* Session) GetOwner() interface{} {
	return this.owner
}
// 设置 owner
func (this* Session) SetOwner(owner interface{}) {
	this.owner = owner
}
// 更新心跳
func (this* Session) updateHeartBeat() {
	this.time = time.Now()
}
