package main

import (
	"github.com/netc0/netco"
	"log"
)

const HELLO_WORLD = "helloworld"

// 发送消息
type HelloWorld struct {
	Name string
}

type GoodBye struct {
	Name string
}


// 启动事件服务
func StartEventService() {
	//var app = gate.Context
	//listener := events.NewEventListener(myHandler)
	//app.EventDispatcher.AddEventListener(HELLO_WORLD, listener)
	//
	//app.EventDispatcher.DispatchEvent(events.NewEvent(HELLO_WORLD, HelloWorld{Name:"Hello"}))
	//app.EventDispatcher.DispatchEvent(events.NewEvent(HELLO_WORLD, GoodBye{Name:"Bye"}))
}

func myHandler(event netco.Event) {
	var hw, ok = event.Object.(HelloWorld)
	if ok {
		log.Println(event.Type, hw, event.Target)
	} else {
		log.Println("event message error", event.Type, event.Object, event.Target)
	}
}