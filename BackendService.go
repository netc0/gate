package main

import (
	"log"
	"github.com/netc0/netco"
)

var (
	proxy *GateProxy
)

func StartBackendService(app *netco.App, config *BackendConfig,
	getSessionCallback func (string)(interface{})) {
	proxy = NewGateProxy(getSessionCallback)
	proxy.AuthCode = config.Auth
	app.SetRPCServerHost(config.Host, proxy)   // 启动 RPC 服务器
}

func BackendServiceDispatch(s interface{}, requestId uint32, routeId uint32, data []byte) {
	session, ok := s.(ISession)
	if !ok {
		return
	}

	msg := netco.RPCGateRequest{}
	msg.RequestId = requestId
	msg.RouteId  = routeId
	msg.Data = data

	msg.ClientId = session.GetId()

	err := DispatchRequest(proxy, msg)
	if err != nil {
		log.Println(err)
		r := netco.BuildSimpleMessage(404, err.Error())
		session.Response(requestId, r)
	}
}