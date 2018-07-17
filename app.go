package main

import (
	"log"
	"github.com/netc0/netco"
	"github.com/netc0/gate/backend"
	"github.com/netc0/gate/frontend"
	"flag"
	"os"
	"github.com/netc0/netco/common"
	"github.com/netc0/gate/web"
)

type GateApp struct {
	netco.App

	frontendConfig frontend.Config
}

// 解析参数
func (this *GateApp) parseArgs() {
	var help bool
	flag.BoolVar(&help, "h", false, "显示帮助")
	// frontend
	this.frontendConfig.TCPBindAddress = *flag.String("t", ":9000", "TCP Bind Address")
	this.frontendConfig.UDPBindAddress = *flag.String("u", ":9001", "UDP Bind Address")
	// backend
	mName := *flag.String("n", "gate", "Node Name")
	mAddress := *flag.String("r", ":9002", "Node Bind Address")

	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}
	this.SetNodeName(mName)
	this.SetNodeAddress(mAddress)
	this.SetGateAddress(mAddress)
}

var (
	logger = common.GetLogger()
)

func (this *GateApp) OnStart() {
	logger.Prefix("[gate] ")
	logger.Debug("网关启动")

	this.parseArgs()

	this.RegisterService("frontend-service", &frontend.Service{App:this})
	this.RegisterService("backend-service", &backend.Service{App:this})
	this.RegisterService("gate-web", &web.WebService{App:this})

	//post config
	this.App.DispatchEvent("frontend.config", this.frontendConfig)
}

func (this *GateApp) OnDestroy() {
	logger.Debug("[gate app] 网关关闭")
}

func NewApp() *GateApp {
	this := &GateApp{}
	//this.Init()
	this.Derived = this
	return this
}


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	app := NewApp()
	app.Run()
}
