package common

import "hash/crc32"

// 状态数据
const (
	// session 状态
	SessionState_Invalid = iota // 无效
	SessionState_Closed		 	// 已关闭
	SessionState_Connected		// 已连接
	SessionState_Disconnected   // 已断开
)

const (
	PacketType_NAN = iota
	PacketType_SYN
	PacketType_ACK
	PacketType_HEARTBEAT
	PacketType_DATA
	PacketType_PUSH
	PacketType_KICK
)

type Packet struct {
	Type uint8
	Body []byte
}

type SessionEvent struct {
	Type int
	Data *Packet
}

func PacketToBinary(Type int, data []byte) []byte{
	if data == nil { data = make([]byte, 0)}
	var result = make([]byte, 4)
	var bodyLen = uint32(len(data))
	result[0] = byte(Type)
	result[1] = byte(bodyLen >> 16)
	result[2] = byte(bodyLen >> 8)
	result[3] = byte(bodyLen >> 0)

	result = append(result, data...)

	return result
}

func PacketResponseToBinary(Type int, requstId uint32, data []byte) []byte{
	if data == nil { data = make([]byte, 0)}
	var result = make([]byte, 8)
	var bodyLen = 4 + uint32(len(data))
	result[0] = byte(Type)
	result[1] = byte(bodyLen >> 16)
	result[2] = byte(bodyLen >> 8)
	result[3] = byte(bodyLen >> 0)

	result[4] = byte(requstId >> 24)
	result[5] = byte(requstId >> 16)
	result[6] = byte(requstId >> 8)
	result[7] = byte(requstId >> 0)

	result = append(result, data...)

	return result
}

func PacketPushToBinary(route string, data []byte) []byte{
	if data == nil { data = make([]byte, 0)}
	var result = make([]byte, 8)
	var bodyLen = 4 + uint32(len(data))

	Type := PacketType_PUSH //推送消息
	var routeId uint32
	routeId = crc32.ChecksumIEEE([]byte(route))

	result[0] = byte(Type)
	result[1] = byte(bodyLen >> 16)
	result[2] = byte(bodyLen >> 8)
	result[3] = byte(bodyLen >> 0)

	result[4] = byte(routeId >> 24)
	result[5] = byte(routeId >> 16)
	result[6] = byte(routeId >> 8)
	result[7] = byte(routeId >> 0)

	result = append(result, data...)

	return result
}