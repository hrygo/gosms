package codec

import (
	"encoding/binary"

	"github.com/hrygo/log"
)

const (
	Pkl              = "pkl" // 数据包总长度
	Cmd              = "op"  // 数据包类型（命令类型）
	Seq              = "seq" // 序号（一对请求与响应序号相同）
	HeadLen   uint32 = 12
	PacketMax        = 5120
)

type Operation interface {
	ToInt() uint32
	OpLog() log.Field
	String() string
}

type Logger interface {
	Log() []log.Field
}

type IHead interface {
	Logger
	Encode() []byte
	Decode([]byte) error
}

type Pdu interface {
	Logger
	Encode() []byte
	Decode(seq uint32, frame []byte) error
}

type RequestPdu interface {
	Pdu
	ToResponse(code uint32) Pdu
}

func UnpackHead(h []byte) (pkl, cmd, seq uint32) {
	if len(h) >= 12 {
		pkl = binary.BigEndian.Uint32(h[0:4])
		cmd = binary.BigEndian.Uint32(h[4:8])
		seq = binary.BigEndian.Uint32(h[8:12])
	}
	return
}
