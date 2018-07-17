package backend

import (
	"github.com/netc0/netco/def"
	"github.com/netc0/netco"
	"github.com/netc0/netco/common"
	"github.com/netc0/netco/rpc"
	"sync"
)

type Service struct {
	def.IService
	App netco.IApp
	rpc.MailHandler

	mailBox rpc.IMailBox

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
		this.onData(obj)
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
		this.App.DispatchEvent("frontend.response", obj)
	})
	// 通知后端, 此连接已经被移除
	this.App.OnEvent("backend.removeSession", func(obj interface{}) {
		this.onRemoveSession(obj)
	})
	go this.startBackend()
}

func (this *Service) OnDestroy() {
	logger.Debug("关闭后端服务")
	this.mailBox.Stop()
}

func (this *Service) onData(obj interface{}) {
	if info, err := def.CastMailClientInfo(obj); err == nil {
		if be := this.getBackend(info.Route); be != nil {
			this.mailBox.SendTo(be.address, &rpc.Mail{Type: def.Mail_RequestData, Object: info})
		} else {
			logger.Debug("没有这个后端", info.Route)
		}
	}
}

// 启动后端服务器
func (this *Service) startBackend() {
	this.mailBox = rpc.NewMailBox(this.App.GetNodeAddress())
	this.mailBox.SetHandler(this)
	go this.mailBox.Start()
}

func (this* Service) reg(obj interface{}) {
	var v def.MailNodeInfo
	switch t := obj.(type) {
	default:
		return
	case def.MailNodeInfo:
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
	this.mailBox.Remove(be.address)
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

// 移除会话
func (this *Service) onRemoveSession(obj interface{}) {
	if info, err := def.CastMailClientInfo(obj); err == nil {
		logger.Debug("移除会话..", info)
		this.mailBox.SendTo(info.RemoteAddress, &rpc.Mail{Type:def.Mail_ClientNotFound, Object:info})
	}
}