package frontend

import (
	"log"
	"net"
	"time"
	"github.com/netc0/netco/def"
	"github.com/netc0/netco"
	"github.com/netc0/gate/protocol"
)

type UDPService struct {
	def.IService
	Transporter

	App netco.IApp
	Config Config
}

func (this *UDPService) OnStart() {
	logger.Debug("启动 UDPService")
	this.mIPFilter = NewIPFilter(true)

	this.App.OnEvent("frontend.udp.restart", func(obj interface{}) {
		switch t:= obj.(type) {
		default:
			return
		case Config:
			this.Config = t
		}
		go this.waitConnection(this.Config.UDPBindAddress)
	})
}
func (this *UDPService) waitConnection(host string) {
	addr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		log.Println("解析 UDP Host 失败", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("启动 UDP 失败", err)
		return
	}

	defer conn.Close()
	logger.Debug("Frontend启动 UDP", host)
	for {
		if err := this.handleClient(conn); err != nil {
			log.Println("UDP 异常", err)
			//break
		}
	}
}

func (this *UDPService) handleClient(conn *net.UDPConn) error {
	data := make([]byte, 2048)
	n, remoteAddr, err := conn.ReadFromUDP(data)

	if err != nil {
		return err
	}

	IPString := remoteAddr.IP.String()
	if !this.mIPFilter.IsAllow(IPString) {
		logger.Debug("禁止 tcp:", IPString)
		return nil
	}

	psession := GetSessionManager().GetSession(remoteAddr.String())
	if psession == nil {
		var session UDPSession
		session.id_int = GetSessionManager().NewSessionId()
		session.OnDataPacket = this.OnDataPacket
		session.time = time.Now() // 更新心跳
		session.isOk = true
		session.conn = conn
		session.remote = remoteAddr
		session.holder = session
		session.id = remoteAddr.String()
		GetSessionManager().AddSession(&session)       // 新增会话
	}
	psession = GetSessionManager().GetSession(remoteAddr.String())
	psession.HandleBytes(data[:n])
	return nil
}

func (this *UDPService) OnDestroy() {
	logger.Debug("关闭 UDPService")
}

func (this *UDPService) OnDataPacket(s protocol.ISession, req uint32, route uint32, data []byte) {
	var v def.MailClientInfo
	v.ClientId = s.GetId()
	v.RequestId = req
	v.Route = route
	v.Data  = data
	v.SourceName = this.App.GetNodeName()
	v.SourceAddress = this.App.GetNodeAddress()
	this.App.DispatchEvent("backend.onData", v)
}
