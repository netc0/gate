package main

import (
	"net"
	"log"
	"time"
)

func (this *TCPTransporter) start() {
	this.running = true
	var l, err = net.Listen("tcp", this.Host)
	if err != nil {
		log.Println(err)
		this.running = false
	}
	log.Println("Frontend启动 TCP", this.Host)
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

func (this *TCPTransporter) handleConnection(conn net.Conn) {
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
			log.Println("读数据错误",err)
			break
		}
		data := buf[:size]

		session.HandleBytes(data)
	}
}
