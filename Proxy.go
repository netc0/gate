package main

import (
	"errors"
	"net/rpc"
	"log"
	"sync"
	"time"
	"fmt"
	"hash/crc32"
	"github.com/netc0/netco"
)

type BackendInfo struct{
	nodeName     string
	client       * rpc.Client
	heatBeatTime time.Time
	a netco.App
}

type GateProxy struct {
	AuthCode string
	Backends map[string]*BackendInfo
	backendLock *sync.Mutex
	getSessionCallback func (string)(interface{})
	backendCache map[uint32]string // crc -> backend name
	backendRoute map[uint32]string // crc -> route
}

//
func NewGateProxy(getSessionCallback func (string)(interface{})) *GateProxy {
	var v GateProxy
	v.init()
	v.getSessionCallback = getSessionCallback
	return &v
}

// 初始化
func (this *GateProxy) init() {
	this.Backends = make(map[string]*BackendInfo)
	this.backendLock = new(sync.Mutex)
	this.backendCache = make(map[uint32]string)
	this.backendRoute = make(map[uint32]string)
	//
	go func() {
		ticker := time.NewTicker(time.Second * 3)
		for range ticker.C {
			go this.checkHeartBeat()
		}
	}()
}

// RPC 注册后端
func (this *GateProxy) RegisterBackend(info netco.RPCBackendInfo, reply* int) error {
	log.Println("RegisterBackend", info)
	if this.AuthCode != info.Auth {
		return errors.New("Auth Code Invalid.")
	}
	return this.addBackend(info)
}

// RPC 回复
func (this *GateProxy) Reply(info netco.RPCGateResponse, reply* int) error {
	return this.reply(info)
}

// RPC Push
func (this *GateProxy) Push(info netco.RPCGatePush, reply* int) error {
	return this.push(info)
}

// 通用消息
func (this *GateProxy) OnMessage(info netco.RPCMessage, reply* int) error {
	return this.onMessage(info)
}

// BackendHeartBeat
func (this *GateProxy) BackendHeartBeat(info netco.RPCBackendInfo, reply* int) error {
	if this.AuthCode != info.Auth {
		return errors.New("Auth Code Invalid.")
	}

	this.backendLock.Lock()
	defer this.backendLock.Unlock()

	var backend = this.getBackend(info.Name)
	if backend == nil {
		return errors.New("Backend not register")
	}
	backend.heatBeatTime = time.Now()
	return nil
}

// 注册后端
func (this *GateProxy) addBackend(info netco.RPCBackendInfo) error {
	// 注册路由
	log.Println("注册后端", info)
	this.backendLock.Lock()
	defer this.backendLock.Unlock()
	if this.Backends[info.Name] != nil {
		delete(this.Backends, info.Name)
	}

	cli, err := netco.RPCClientConnect(info.RCPRemote)
	if err != nil {
		return err
	}
	var backend BackendInfo
	this.Backends[info.Name] = &backend
	backend.client = cli
	backend.heatBeatTime = time.Now()
	backend.nodeName = info.Name
	// 注册 cache
	for _, route := range info.Routes {
		var crc = crc32.ChecksumIEEE([]byte(route))
		this.backendRoute[crc] = route
		this.backendCache[crc] = info.Name
	}
	return nil
}

// 回复客户端
func (this *GateProxy) reply(info netco.RPCGateResponse) error {
	var s = this.getSession(info.ClientId)
	session, ok := s.(ISession)

	if session == nil || ok == false {
		return errors.New("Client session not found.")
	}
	//session.Response(connector.PacketType_DATA, info.RequestId, info.Data)
	session.Response(info.RequestId, info.Data)
	return nil

}
func (this *GateProxy) push(info netco.RPCGatePush) error {
	var s = this.getSession(info.ClientId)
	session, ok := s.(ISession)

	if session == nil || ok == false {
		return errors.New("Client session not found.")
	}
	if session == nil {
		return errors.New("Client session not found.")
	}

	session.Push(info.Data)

	return nil
}


// 获取后端
func (this *GateProxy) getBackend(name string) *BackendInfo {
	return this.Backends[name]
}
// 获取客户端
func (this *GateProxy) getSession(name string) interface{} {
	return this.getSessionCallback(name)
}

// 心跳检测
func (this *GateProxy) checkHeartBeat() {
	this.backendLock.Lock()
	defer this.backendLock.Unlock()
	for k, backend := range this.Backends {
		if time.Now().Second() - backend.heatBeatTime.Second() > 5 {
			log.Println("Backend heartbeat fail:", backend)
			delete(this.Backends, k)

		}
	}
}
// 根据CRC获取路由, 返回 routeName, 后端名称
func (this *GateProxy) crcBackend(crc uint32) (string, *BackendInfo) {
	var backendName = this.backendCache[crc]
	if backendName == "" {
		return "", nil
	}
	return this.backendRoute[crc], this.Backends[backendName]
}

func DispatchRequest(this *GateProxy, msg netco.RPCGateRequest) error {
	var routeName, backend = this.crcBackend(msg.RouteId)
	if routeName == "" || backend == nil  {
		return errors.New("Backend not exist")
	}

	rs := backend.client.Call(routeName, msg, nil)
	if rs != nil {
		return errors.New(fmt.Sprintf("Backend Invoke Error:%v", rs.Error()))
	}

	return nil
}

func (this *GateProxy) onMessage(msg netco.RPCMessage) error {
	if msg.AuthCode != this.AuthCode {
		return errors.New("AuthCode error")
	}
	if msg.Command == 1 { // 会话关闭时需要通知我
		sid := string(msg.Value)
		log.Println("如果", sid, "关闭了 请告诉我", msg.ResponseNodeName)

		session := GetSession(sid)
		if session != nil {
			session.AddCloseEventListener(func(session ISession) {
				log.Println("会话" + sid + "关闭啦")
				backend := this.getBackend(msg.ResponseNodeName)
				if backend != nil {
					var req netco.RPCMessage
					req.Value = []byte(sid)
					req.AuthCode = msg.ResponseAuthCode
					var r int
					backend.client.Call(msg.ResponseRoute, req, &r)
				}
			})
		}
	}
	return nil
}