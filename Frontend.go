package main

import (
	"sync"
	"log"
)

// define

func (this* TCPSession) IsOk() bool {
	return this.isOk
}

var (
	sessions map[string]ISession
	sessionMutex *sync.Mutex
	DispatchBackendCallback func(s interface{}, requestId uint32, routeId uint32, data []byte)

	onNewSession func(sid string)
	onCloseSession func(sid string)
)

type TCPTransporter struct {
	Transporter
}

type UDPTransporter struct {
	Transporter
}

// 启动前端服务
func StartFrontendSerice(config* FrontendConfig) {
	sessions = make(map[string]ISession)
	sessionMutex = new(sync.Mutex)

	if config.TCPHost != "" {
		var tcp TCPTransporter
		tcp.Host = config.TCPHost
		tcp.OnDataPacket = func(s interface{}, requestId uint32, routeId uint32, data []byte) {
			DispatchBackend(s, requestId, routeId, data)
		}

		go func() {
			tcp.start()
		}()
	}

	if config.UDPHost != "" {
		var udp UDPTransporter
		udp.Host = config.UDPHost
		udp.OnDataPacket = func(s interface{}, requestId uint32, routeId uint32, data []byte) {
			DispatchBackend(s, requestId, routeId, data)
		}

		go func() {
			udp.start()
		}()
	}
}

// 设置session 转发回调
func SetDispatchBackendCallback(callback func(session interface{}, requestId uint32, routeId uint32, data []byte)) {
	DispatchBackendCallback = callback
}

func GetSession(sid string) ISession {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	var s = sessions[sid]
	return s
}

// 清空会话
func ClearSession(owner interface{}) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	for k, v := range sessions {
		if v.GetOwner() == owner {
			delete(sessions, k)
		}
	}
}

// 遍历会话
func ForeachSession(callback func(session ISession)) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	for _, v := range sessions {
		callback(v)
	}
}
// 添加会话Frontend启动
func AddSession(inst interface{}) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	session, ok := inst.(ISession)
	if ok {
		log.Println("新连接进入", session.GetId(), session.GetOwner())
		sessions[session.GetId()] = session
		//log.Println("cast to ISession ok")
		return
	}

	log.Println("cast to ISession error")
}
// 删除会话
func RemoveSession(inst interface{}) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	session, ok := inst.(ISession)
	if ok {
		log.Println("Frontend 关闭会话", session.GetId())
		delete(sessions, session.GetId())
	}
}

// 传递消息到后端
func DispatchBackend(s interface{}, requestId uint32, routeId uint32, data []byte) {
	//log.Println("收到数据:",requestId, routeId, string(data))
	DispatchBackendCallback(s, requestId, routeId, data)
}