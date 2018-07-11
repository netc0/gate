package frontend

import (
	"log"
	"net"
	"time"
	"github.com/netc0/netco/def"
	"github.com/netc0/netco"
	"github.com/netc0/gate/common"
	"github.com/netc0/gate/models"
)

type UDPService struct {
	def.IService
	Transporter

	App netco.IApp
	Config Config
}

func (this *UDPService) OnStart() {
	logger.Debug("启动 UDPService")

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
	//log.Println("Frontend启动 UDP:", host)
	defer conn.Close()

	for {
		this.handleClient(conn)
	}
}

func (this *UDPService) handleClient(conn *net.UDPConn) {
	data := make([]byte, 2048)
	n, remoteAddr, err := conn.ReadFromUDP(data)

	if err != nil {
		log.Println(err)
		return
	}

	psession := GetSession(remoteAddr.String())
	if psession == nil {
		var session UDPSession
		session.OnDataPacket = this.OnDataPacket
		session.time = time.Now() // 更新心跳
		session.isOk = true
		session.conn = conn
		session.remote = remoteAddr
		session.holder = session
		session.id = remoteAddr.String()
		AddSession(&session)       // 新增会话
	}
	psession = GetSession(remoteAddr.String())
	psession.HandleBytes(data[:n])
}

func (this *UDPService) OnDestroy() {
	logger.Debug("关闭 UDPService")
}

func (this *UDPService) OnDataPacket(s common.ISession, req uint32, route uint32, data []byte) {
	i := models.FrontendRequestInfo{Session:s, RequestId:req, Route:route, Data:data}
	this.App.DispatchEvent("backend.onData", i)
}