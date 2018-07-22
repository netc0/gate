package web

import (
	"net/http"
	"encoding/json"
	"github.com/netc0/gate/frontend"
	"bytes"
	"fmt"
)

func responseJson(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if data, err := json.Marshal(v); err == nil {
		w.Write(data)
	}
}

func responseText(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/text")
	w.Write([]byte(msg))
}

func responseByte(w http.ResponseWriter, data []byte) {
	w.Write(data)
}

func api_home(w http.ResponseWriter, req* http.Request) {
	responseText(w, "netco网关接口")
}

func api_getSession(w http.ResponseWriter, req* http.Request) {
	result := frontend.GetSessionManager().API_getSession()
	responseJson(w, result)
}

func api_listAPI(w http.ResponseWriter, req* http.Request) {
	var buffer bytes.Buffer
	for k, _ := range CurrentService.webRoute {
		buffer.WriteString(fmt.Sprintf("<a herf=\"%v\">%v</a><br>", k, k))
	}
	logger.Debug(string(buffer.Bytes()))
	responseByte(w, buffer.Bytes())
}