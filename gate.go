package main

import (
    "github.com/netc0/netco"
    "log"
    "flag"
    "os"
)


type AppArgs struct  {
    help    bool
    RPCAuth string
    RPCHost string

    TCPHost string
    UDPHost string
}

var (
    appArgs AppArgs
)

func parseArgs() {
    flag.BoolVar(&appArgs.help, "h", false, "显示帮助")
    flag.StringVar(&appArgs.RPCAuth, "k", "netc0", "RPC 验证码")
    flag.StringVar(&appArgs.RPCHost, "r", ":9002", "RPC Host")

    flag.StringVar(&appArgs.TCPHost, "t", ":9000", "TCP Host")
    flag.StringVar(&appArgs.UDPHost, "u", ":9001", "TCP Host")
    flag.Parse()
}

func processArgs() {
    if appArgs.help {
        flag.Usage()
        os.Exit(0)
    }
}

func setupFrontend(config* FrontendConfig) {
    config.TCPHost = ":9000"
    config.UDPHost = ":9001"
}
func setupBackend(config* BackendConfig) {
    config.Host = appArgs.RPCHost
    config.Auth = appArgs.RPCAuth
}


func startApp () {
    var context = netco.NewApp()

    // 前端配置参数
    var frontendConfig FrontendConfig
    setupFrontend(&frontendConfig)

    // 后端配置
    var backendConfig BackendConfig
    setupBackend(&backendConfig)

    StartFrontendSerice(&frontendConfig) // 前端服务
    SetDispatchBackendCallback(DispatchBackendFunc)

    StartBackendService(&context, &backendConfig, getSessionCallback) // 后端服务
    context.Start()
}

func DispatchBackendFunc(s interface{}, requestId uint32, routeId uint32, data []byte) {
    BackendServiceDispatch(s, requestId, routeId, data);
}

func getSessionCallback(sid string) interface{}{
    return GetSession(sid)
}


func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    parseArgs()   // 解析参数
    processArgs() // 处理参数
    startApp()    // 启动
}