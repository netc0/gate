package protocol

type PacketReader struct {
	buffer    []byte
	readType  int
	pkgType   uint8
	pkgLength int
}

const Read_Header = 0
const Read_Body = 1

func (reader*PacketReader) ParsePacket(newData []byte) *Packet {
	if newData != nil {
		reader.buffer = append(reader.buffer, newData[:]...)
	}
	if reader.readType == Read_Header {
		if len(reader.buffer) < 4 {
			return nil
		}
		// 解析header
		reader.pkgType = uint8(reader.buffer[0])
		reader.pkgLength = int(
				reader.buffer[1] << 16 |
				reader.buffer[2] << 8 |
				reader.buffer[3]);
		reader.buffer = reader.buffer[4:]
		reader.readType = Read_Body

		return reader.ParsePacket(nil) // continue parse
	} else if reader.readType == Read_Body {
		if len(reader.buffer) >= reader.pkgLength { // 解析 body 完毕
			var np Packet
			np.Type = reader.pkgType
			if reader.pkgLength > 0 {
				np.Body = reader.buffer[:reader.pkgLength]
				reader.buffer = reader.buffer[reader.pkgLength:]
			}

			reader.pkgLength = 0
			reader.readType = Read_Header

			return &np
		}
	}
	return nil
}
