package web

import (
	"github.com/netc0/netco/def"
	"github.com/netc0/netco/common"
	"github.com/netc0/netco"
	"net/http"
	"context"
	"time"
	"fmt"
)

type WebService struct {
	def.IService
	http.Handler

	App netco.IApp

	webServer *http.Server
	webRoute map[string] func(response http.ResponseWriter, request *http.Request)
}
var (
	logger = common.GetLogger()
	CurrentService *WebService
)

func (this *WebService) OnStart() {
	logger.Prefix("[web-service] ")
	logger.Debug("启动 web 服务")
	CurrentService = this

	this.webServer = &http.Server{Addr:":9090", Handler:this}
	this.webRoute = make(map[string]func(response http.ResponseWriter, request *http.Request))
	// 注册路由
	this.webRoute["/"] = api_home
	this.webRoute["/get_session"] = api_getSession
	this.webRoute["/list_api"] = api_listAPI
	http.NewServeMux()

	go func() {
		if err := this.webServer.ListenAndServe(); err != nil {
			logger.Debug(err)
		}
	}()
}
func (this *WebService) OnDestroy() {
	logger.Debug("关闭 web 服务")
	ctx, _ := context.WithTimeout(context.Background(), 5* time.Second)
	if err := this.webServer.Shutdown(ctx); err != nil {
		logger.Debug(err)
	}
}

func (this *WebService) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	// TODO 验证请求中的 token

	// ...
	if cb := this.webRoute[request.URL.Path]; cb != nil{
		cb(response, request)
	} else {
		response.WriteHeader(404)
		response.Write([]byte(fmt.Sprintf("page(%v) nout found", request.URL.Path)))
	}
}
