package frontend

import (
	"net"
	"time"
	"log"
	"fmt"
	"github.com/netc0/gate/protocol"
	"sync"
	"sync/atomic"
)

type Session struct {
	protocol.ISession
	holder       interface{}
	id           string
	id_int       int32
	isOk         bool
	reader       protocol.PacketReader
	time         time.Time // 心跳
	OnDataPacket func(protocol.ISession, uint32, uint32, []byte)

	owner interface{}

	closeEventListeners []func(protocol.ISession)
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

type SessionManager struct {
	sessionId int32
	sessions map[string]protocol.ISession
	sessionMutex *sync.Mutex
}

type SessionInfo struct {
	Id            int32  `json:"id"`
	RemoteAddress string `json:"remote"`
	Type          string `json:"type"`
}

var (
	gSessionManager SessionManager
)

// 获取 ID
func (this *Session)GetId() string { return this.id }
// 设置 ID
func (this *Session) SetId(id string) { this.id = id }
// 获取 ID
func (this *Session)GetIdInt32() int32 { return this.id_int }
// 设置 ID
func (this *Session) SetIdInt32(id int32) { this.id_int = id }
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
func (this *Session)Response(requestId, statusCode uint32, r[]byte){
	var data = protocol.PacketResponseToBinary(protocol.PacketType_DATA, requestId, statusCode, r)
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
		logger.Debug("close tcp conn")
		t.conn.Close()
		t.conn = nil
	case UDPSession:
		logger.Debug("close udp conn")
		t.conn = nil // 不能close 只能赋值为空
		break
	}

	for _, callback := range this.closeEventListeners {
		callback(this)
	}

	GetSessionManager().RemoveSession(this)
}
// 状态是否正常
func (this *Session)IsOk() bool{ return false }
// 处理数据包
func (this *Session)HandlePacket(packet protocol.Packet) int {
	this.time = time.Now()
	if packet.Type == protocol.PacketType_SYN { // 收到 SYN
		var data = protocol.PacketToBinary(protocol.PacketType_ACK, nil)
		this.send(data) // 必须回应SYN
		return 0
	} else if packet.Type == protocol.PacketType_ACK { // 收到 ACK
		return 0
	} else if packet.Type == protocol.PacketType_HEARTBEAT { // 纯心跳包 一般不需要
		return 0
	} else if packet.Type == protocol.PacketType_DATA { // on data
		this.onDataPacket(packet.Body)
		return 0
	} else if packet.Type == protocol.PacketType_KICK { // on kick

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
func (this*Session) AddCloseEventListener(callback func(session protocol.ISession)) {
	this.closeEventListeners = append(this.closeEventListeners, callback)
}
// 发送 TCP 消息
func (this* TCPSession) send(data[]byte) {
	b, err := this.conn.Write(data)
	if err != nil {
		log.Println("tcp write: ", b, err)
	}
}
// TCP isOK
func (this *TCPSession)IsOk() bool{
	if this.conn == nil {
		return false
	}

	return true
}
// UDP isOK
func (this *UDPSession)IsOk() bool{
	if this.conn == nil {
		return false
	}

	return true
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
// ToString
func (this* Session)ToString() string {
	return fmt.Sprintf("DefaultSession")
}
// ToString
func (this* TCPSession)ToString() string {
	return fmt.Sprintf("TCPSession")
}
// ToString
func (this* UDPSession)ToString() string {
	return fmt.Sprintf("UDPSession")
}


// 获取 SessionManager 实例
func GetSessionManager() *SessionManager {
	return &gSessionManager
}
// 初始化 SessionManager
func (this *SessionManager) Init() {
	this.sessions = make(map[string]protocol.ISession)
	this.sessionMutex = new(sync.Mutex)
}
// 获取会话
func (this *SessionManager) GetSession(sid string) protocol.ISession {
	this.sessionMutex.Lock()
	result := this.sessions[sid]
	this.sessionMutex.Unlock()
	return result
}

// 新增会话
func (this *SessionManager) AddSession(s protocol.ISession) {
	this.sessionMutex.Lock()
	this.sessions[s.GetId()] = s
	this.sessionMutex.Unlock()
}

// 清空会话
func (this *SessionManager) ClearSession(owner interface{}) {
	this.sessionMutex.Lock()
	defer this.sessionMutex.Unlock()
	for k, v := range this.sessions {
		if v.GetOwner() == owner {
			delete(this.sessions, k)
		}
	}
}

// 遍历会话
func (this *SessionManager) ForeachSession(callback func(session protocol.ISession)) {
	this.sessionMutex.Lock()
	defer this.sessionMutex.Unlock()
	for _, v := range this.sessions {
		callback(v)
	}
}

// 删除会话
func (this *SessionManager) RemoveSession(session protocol.ISession) {
	this.sessionMutex.Lock()
	defer this.sessionMutex.Unlock()
	delete(this.sessions, session.GetId())
}
// 生成会话 ID
func (this *SessionManager) NewSessionId() int32 {
	atomic.AddInt32(&this.sessionId, 1)
	return this.sessionId
}

// 获取会话信息
func (this *SessionManager) API_getSession() []SessionInfo{
	var ss []SessionInfo
	GetSessionManager().ForeachSession(func(session protocol.ISession) {
		sid := session.GetId()
		id := session.GetIdInt32()
		s := SessionInfo{RemoteAddress:sid, Id:id, Type:session.ToString()}
		ss = append(ss, s)
	})
	return ss
}
