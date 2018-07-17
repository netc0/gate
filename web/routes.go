package web

import (
	"net/http"
	"encoding/json"
)

func responseJson(w http.ResponseWriter, v interface{}) {
	if data, err := json.Marshal(v); err == nil {
		w.Write(data)
	}
}

func route_home(r http.ResponseWriter, req* http.Request) {
	r.Write([]byte("index"))
}

func route_echo(r http.ResponseWriter, req* http.Request) {
	r.Write([]byte("echo"))
}

func route_get_session(w http.ResponseWriter, req* http.Request) {
	CurrentService.App.DispatchEvent("frontend.web.get_session", func(obj interface{}) {
		 responseJson(w, obj)
	})
}