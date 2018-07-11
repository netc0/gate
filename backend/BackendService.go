package backend

import (
	"github.com/netc0/netco/def"
	"github.com/netc0/netco"
	"github.com/netc0/netco/common"
	"github.com/netc0/gate/models"
	"github.com/netc0/netco/rpc"
	"sync"
)

type Service struct {
	def.IService
	App netco.IApp
	rpc.MailHandler

	mailBox rpc.IMailBox

	Config Config
	backends map[string]*Backend
	backendMutex *sync.Mutex

	routeCache map[uint32]*Backend
}

var (
	logger = common.GetLogger()
)

func (this *Service) OnStart() {
	logger.Prefix("[backend] ")
	this.backends = make(map[string]*Backend)
	this.routeCache = make(map[uint32]*Backend)
	this.backendMutex = new(sync.Mutex)

	this.App.OnEvent("backend.onData", func(obj interface{}) {
		switch t := obj.(type) {
		default:
			return 
		case models.FrontendRequestInfo:
			this.onData(t)
		}
	})
	// 注册后端节点
	this.App.OnEvent("backend.reg", func(obj interface{}) {
		this.reg(obj)
	})
	// 注册后端的路由
	this.App.OnEvent("backend.addRoute", func(obj interface{}) {
		this.addRoute(obj)
	})
	// 回复客户端
	this.App.OnEvent("backend.response", func(obj interface{}) {
		this.response(obj)
	})
	go this.startBackend()
}

func (this *Service) OnDestroy() {
	logger.Debug("关闭后端服务")
	this.mailBox.Stop()
}

func (this *Service) onData(i models.FrontendRequestInfo) {
	if be := this.getBackend(i.Route); be != nil {
		var req def.MailClientData
		req.IsRequest = true
		req.ClientId = i.Session.GetId()
		req.RequestId = i.RequestId
		req.Route = i.Route
		req.Data = i.Data
		this.mailBox.SendTo(be.address, &rpc.Mail{Type:def.Mail_RequestData, Object:req})
	} else {
		logger.Debug("没有这个后端", i.Route)
	}
}

// 启动后端服务器
func (this *Service) startBackend() {
	logger.Debug("后端server", this.Config.RPCBindAddress)
	this.mailBox = rpc.NewMailBox(this.Config.RPCBindAddress)
	this.mailBox.SetHandler(this)
	go this.mailBox.Start()
}

func (this* Service) reg(obj interface{}) {
	var v def.MailOffice
	switch t := obj.(type) {
	default:
		return
	case def.MailOffice:
		v = t
	}

	if v.Name == "" || v.Address == ""{
		return
	}

	this.backendMutex.Lock()
	defer this.backendMutex.Unlock()
	var be Backend
	be.name = v.Name
	be.address = v.Address
	this.backends[be.name] = &be

	if err := this.mailBox.Connect(be.address); err != nil {
		logger.Debug("注册后端失败:", err, v)
		return
	}

	logger.Debug("后端注册成功", v)
}

func (this *Service) addRoute(obj interface{}) {
	var v def.MailRoutineInfo
	switch t := obj.(type) {
	default:
		return
	case def.MailRoutineInfo:
		v = t
	}

	this.backendMutex.Lock()
	defer this.backendMutex.Unlock()
	if be := this.backends[v.Name]; be != nil {
		be.routes = v.Routes
		for _,v := range be.routes {
			this.routeCache[v] = be
		}
		logger.Debug("注册路由:", be)
		return
	}
	logger.Debug("注册路由失败, 没有:", v.Name, "后端")
}

func (this *Service) getBackend(route uint32) *Backend {
	return this.routeCache[route]
}

func (this *Service) response(obj interface{}) {
	var resp def.MailClientData
	switch t := obj.(type) {
	default:
		return
	case def.MailClientData:
		resp = t
	}

	var info models.BackendResponseInfo
	info.SessionId = resp.ClientId
	info.RequestId = resp.RequestId
	info.Data      = resp.Data
	this.App.DispatchEvent("frontend.response", info)
}