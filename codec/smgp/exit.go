package smgp

import (
	"github.com/hrygo/log"

	"github.com/hrygo/gosms/codec"
)

type Exit MessageHeader
type ExitRsp MessageHeader

func NewExit(seq uint32) *Exit {
	at := &Exit{PacketLength: codec.HeadLen, RequestId: SMGP_EXIT, SequenceId: seq}
	return at
}

func (t *Exit) Encode() []byte {
	return (*MessageHeader)(t).Encode()
}

func (t *Exit) Decode(seq uint32, _ []byte) error {
	t.PacketLength = codec.HeadLen
	t.RequestId = SMGP_EXIT
	t.SequenceId = seq
	return nil
}

func (t *Exit) ToResponse(_ uint32) codec.Pdu {
	resp := ExitRsp{}
	resp.PacketLength = t.PacketLength
	resp.RequestId = SMGP_EXIT_RESP
	resp.SequenceId = t.SequenceId
	return &resp
}

func (t *Exit) Log() []log.Field {
	return (*MessageHeader)(t).Log()
}

func (t *Exit) String() string {
	return (*MessageHeader)(t).String()
}

func (r *ExitRsp) Encode() []byte {
	return (*MessageHeader)(r).Encode()
}

func (r *ExitRsp) Decode(seq uint32, _ []byte) error {
	r.PacketLength = codec.HeadLen
	r.RequestId = SMGP_EXIT_RESP
	r.SequenceId = seq
	return nil
}
func (r *ExitRsp) Log() []log.Field {
	return (*MessageHeader)(r).Log()
}

func (r *ExitRsp) String() string {
	return (*MessageHeader)(r).String()
}
