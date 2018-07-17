package frontend

import (
	"github.com/netc0/netco/def"
	"github.com/netc0/netco"
	"github.com/netc0/netco/common"
	"github.com/netc0/gate/protocol"
)

type Service struct {
	def.IService

	App netco.IApp
	config Config
}

var (
	logger = common.GetLogger()
)
func (this *Service) OnStart() {
	logger.Prefix("[frontend] ")
	logger.Debug("启动前端服务")
	this.App.OnEvent("frontend.config", func(obj interface{}) {
		switch t:= obj.(type) {
		default:
		case Config:
			this.config = t
			this.onConfig()
		}
	})
	// 回复客户端
	this.App.OnEvent("frontend.response", func(obj interface{}) {
		this.response(obj)
	})

	// 推送消息客户端
	this.App.OnEvent("frontend.push", func(obj interface{}) {
		this.push(obj)
	})

	// 推送消息客户端
	this.App.OnEvent("frontend.web.get_session", func(obj interface{}) {
		this.get_session(obj)
	})

	// 启动 TCP 服务
	this.App.RegisterService("frontend-tcp", &TCPService{App:this.App})
	this.App.RegisterService("frontend-udp", &UDPService{App:this.App})
}

func (this *Service) OnDestroy() {
	logger.Debug("[frontend] 关闭前端服务")
}

// 配置改变
func (this *Service) onConfig() {
	logger.Debug("[frontend] 配置改变", this.config)
	this.App.DispatchEvent("frontend.tcp.restart", this.config)
	this.App.DispatchEvent("frontend.udp.restart", this.config)
}

// 回复客户端
func (this *Service) response(obj interface{}) {
	if resp, err := def.CastMailClientInfo(obj); err == nil {
		if s := GetSession(resp.ClientId); s != nil {
			s.Response(resp.RequestId, resp.Data)
		} else {
			logger.Debug("客户端不存在", resp.ClientId)
		}
	}
}

// 回复客户端
func (this *Service) push(obj interface{}) {
	data, err := def.CastMailClientInfo(obj)
	if err != nil {
		logger.Debug(err)
		return
	}

	if s := GetSession(data.ClientId); s != nil {
		// 构造推送消息
		raw := protocol.PacketPushToBinary(data.Route, data.Data)
		//logger.Debug("推送消息", raw)
		s.Push(raw)
		return
	}
	logger.Debug("push 客户端不存在", data.ClientId)
	var sdata def.MailClientInfo
	sdata.RemoteAddress = data.SourceAddress
	sdata.ClientId = data.ClientId
	this.App.DispatchEvent("backend.removeSession", sdata)
}

// 获取会话信息
func (this *Service) get_session(callback interface{}) {
	switch t := callback.(type) {
	case func(interface{}):
		var ss []SessionInfo
		ForeachSession(func(session protocol.ISession) {
			sid := session.GetId()
			id := session.GetIdInt32()
			s := SessionInfo{SessionId:sid, Id:id}
			ss = append(ss, s)
		})
		t(ss)
	}
}
