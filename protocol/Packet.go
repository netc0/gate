package protocol

// 状态数据

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

func PacketResponseToBinary(Type int, requstId, statusCode uint32, data []byte) []byte{
	if data == nil { data = make([]byte, 0)}
	headerSize := 12
	var result = make([]byte, headerSize)
	var bodyLen = uint32(headerSize) + uint32(len(data)) - 4
	// type and size
	result[0] = byte(Type)
	result[1] = byte(bodyLen >> 16)
	result[2] = byte(bodyLen >> 8)
	result[3] = byte(bodyLen >> 0)
	// requestId
	result[4] = byte(requstId >> 24)
	result[5] = byte(requstId >> 16)
	result[6] = byte(requstId >> 8)
	result[7] = byte(requstId >> 0)
	// statusCode
	result[8]  = byte(statusCode >> 24)
	result[9]  = byte(statusCode >> 16)
	result[10] = byte(statusCode >> 8)
	result[11] = byte(statusCode >> 0)

	result = append(result, data...)

	return result
}

func PacketPushToBinary(routeId uint32, data []byte) []byte{
	if data == nil { data = make([]byte, 0)}
	var result = make([]byte, 8)
	var bodyLen = 4 + uint32(len(data))

	Type := PacketType_PUSH //推送消息

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