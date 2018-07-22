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
	GetSessionManager().Init() // 会话管理器
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
		if s := GetSessionManager().GetSession(resp.ClientId); s != nil {
			logger.Debug("response:", resp.StatusCode, resp.Data)
			s.Response(resp.RequestId, resp.StatusCode, resp.Data)
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

	if s := GetSessionManager().GetSession(data.ClientId); s != nil {
		// 构造推送消息
		raw := protocol.PacketPushToBinary(data.Route, data.Data)
		s.Push(raw)
		return
	}

	var sdata def.MailClientInfo
	sdata.RemoteAddress = data.SourceAddress
	sdata.ClientId = data.ClientId
	this.App.DispatchEvent("backend.removeSession", sdata)
	logger.Debug("push 客户端不存在", data.ClientId, data.SourceAddress)
}
