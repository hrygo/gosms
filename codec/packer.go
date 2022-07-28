package codec

import (
	"encoding/binary"

	"github.com/hrygo/log"
)

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
