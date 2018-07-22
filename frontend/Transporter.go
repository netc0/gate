package frontend

import (
	"log"
	"github.com/netc0/gate/protocol"
)

// 传输接口
type ITransporter interface {
	start()
	releaseSessions()
	checkHeartBeat()
}

// 传输基类
type Transporter struct {
	ITransporter
	running bool   // 是否在运行中
	OnNewConnection func(interface{})
}


func (this *Transporter) releaseSessions(){
	GetSessionManager().ClearSession(this)
}

func (this *Transporter) checkHeartBeat() {
	var die []protocol.ISession
	GetSessionManager().ForeachSession(func(s protocol.ISession) {
		if s.IsTimeout() {
			die = append(die, s)
		}
	})

	for _, s := range die{
		log.Println("session:", s.GetId(), "失去心跳")
		s.Kick()  // 踢下线
		s.Close() //关闭
	}
}
