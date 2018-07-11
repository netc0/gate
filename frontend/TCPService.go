package frontend

import (
	"net"
	"log"
	"time"
	"github.com/netc0/netco/def"
	"github.com/netc0/netco"
	"io"
	"github.com/netc0/gate/common"
	"github.com/netc0/gate/models"
)

type TCPService struct {
	def.IService
	Transporter

	App netco.IApp
	Config Config
}

func (this *TCPService) OnStart() {
	logger.Debug("启动 TCPService")

	this.App.OnEvent("frontend.tcp.restart", func(obj interface{}) {
		switch t:= obj.(type) {
		default:
			return
		case Config:
			this.Config = t
		}
		go this.waitConnection(this.Config.TCPBindAddress)
	})
}

func (this *TCPService) OnDestroy() {
	logger.Debug("关闭 TCPService")
}

// 等待连接
func (this *TCPService) waitConnection(host string) {
	this.running = true
	var l, err = net.Listen("tcp", host)
	if err != nil {
		log.Println(err)
		this.running = false
		return
	}
	logger.Debug("Frontend启动 TCP", host)
	defer l.Close()
	defer log.Println("Close TCP Server")
	defer this.releaseSessions()

	// heart beat service
	go func() {
		var heartBeatService = time.NewTicker(time.Second)
		for range heartBeatService.C {
			go this.checkHeartBeat()
		}
	}()

	for {
		if this.running == false {
			break;
		}
		var conn, err = l.Accept()
		if err != nil {
			break
		}
		go this.handleConnection(conn)
	}
}

// 处理连接
func (this *TCPService) handleConnection(conn net.Conn) {
	var session TCPSession

	defer conn.Close()
	defer RemoveSession(session)

	session.OnDataPacket = this.OnDataPacket
	session.time = time.Now() // 更新心跳
	session.isOk = true
	session.conn = conn
	session.holder = session
	session.id = conn.RemoteAddr().String()
	AddSession(&session)       // 新增会话

	for {
		if !session.IsOk() {
			log.Println("break")
			break
		}
		buf := make([]byte, 1024)
		size, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				logger.Debug("读数据错误", err)
			}
			session.Close()
			break
		}
		data := buf[:size]

		session.HandleBytes(data)
	}
}

func (this *TCPService) OnDataPacket(s common.ISession, req uint32, route uint32, data []byte) {
	i := models.FrontendRequestInfo{Session:s, RequestId:req, Route:route, Data:data}
	this.App.DispatchEvent("backend.onData", i)
}