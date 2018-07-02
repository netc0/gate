package main

import "log"

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
	Host    string // 绑定的Host
	OnNewConnection func(interface{})
	OnDataPacket func(interface{}, uint32, uint32, []byte) // 收到消息
}


func (this *Transporter) releaseSessions(){
	ClearSession(this)
}

func (this *Transporter) checkHeartBeat() {
	var die []ISession
	ForeachSession(func(s ISession) {
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

