package backend

import (
	"github.com/netc0/netco/rpc"
	"github.com/netc0/netco/def"
)

func (this *Service) OnNewMail(mail rpc.Mail) {
	if mail.Type == 0 {
		// 心跳包
		return
	} else if mail.Type == def.Mail_Reg {
		// 注册
		var v def.MailOffice
		if err := mail.Decode(&v); err == nil {
			this.App.DispatchEvent("backend.reg", v)
		}
	} else if mail.Type == def.Mail_AddRoute {
		var v def.MailRoutineInfo
		if err := mail.Decode(&v); err == nil {
			this.App.DispatchEvent("backend.addRoute", v)
		}
	} else if mail.Type == def.Mail_ResponseData {
		var v def.MailClientData
		if err := mail.Decode(&v); err != nil {
			logger.Debug(err)
			return
		}
		this.App.DispatchEvent("backend.response", v)
	}
}

func (this *Service)OnRoutineConnected(remote string) {
	logger.Debug("连接到节点:", remote)
}
func (this *Service)OnRoutineDisconnect(remote string, err error) {
	logger.Debug("节点断开:", remote, err)
}
