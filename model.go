package main

import "net/rpc"

type BackendConfig struct {
	Host string
	Auth string
}

type FrontendConfig struct {
	TCPHost string
	UDPHost string
}

type GateRPCRecord struct {
	remote string
	client *rpc.Client
	routes []string
}
