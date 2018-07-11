package models

import "github.com/netc0/gate/common"

type FrontendRequestInfo struct {
	Session   common.ISession
	RequestId uint32
	Route     uint32
	Data      []byte
}

type BackendResponseInfo struct {
	SessionId string
	RequestId uint32
	Data      []byte
}