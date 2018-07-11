package frontend

import (
	"github.com/netc0/netco/def"
	"github.com/netc0/netco"
	"github.com/netc0/netco/common"
	"github.com/netc0/gate/models"
	"errors"
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
	var resp models.BackendResponseInfo
	switch t:= obj.(type) {
	default:
		return
	case models.BackendResponseInfo:
		resp = t
	}
	if s := GetSession(resp.SessionId); s != nil {
		s.Response(resp.RequestId, resp.Data)
	} else {
		logger.Debug("客户端不存在", resp.SessionId)
	}
}

func GetClientData(obj interface{}) (def.MailClientData, error) {
	var info def.MailClientData
	switch t:= obj.(type) {
	default:
		return info, errors.New("cast to MailClientData failed")
	case def.MailClientData:
		info = t
	}
	return info, nil
}