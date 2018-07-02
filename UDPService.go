package main

import (
	"log"
	"net"
	"time"
)

func (this *UDPTransporter) start() {

	addr, err := net.ResolveUDPAddr("udp", this.Host)
	if err != nil {
		log.Println("解析 UDP Host 失败", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("启动 UDP 失败", err)
		return
	}
	log.Println("Frontend启动 UDP:", this.Host)
	defer conn.Close()

	for {
		this.handleClient(conn)
	}
}

func (this *UDPTransporter) handleClient(conn *net.UDPConn) {
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
	psession.updateHeartBeat() // 更新心跳
}
