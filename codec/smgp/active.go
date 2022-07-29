package smgp

import (
	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/codec"
)

type ActiveTest MessageHeader
type ActiveTestRsp MessageHeader

func NewActiveTest() *ActiveTest {
	at := &ActiveTest{PacketLength: codec.HeadLen, RequestId: SMGP_ACTIVE_TEST, SequenceId: uint32(codec.B32Seq.NextVal())}
	return at
}

func (t *ActiveTest) Encode() []byte {
	return (*MessageHeader)(t).Encode()
}

func (t *ActiveTest) Decode(seq uint32, _ []byte) error {
	t.PacketLength = codec.HeadLen
	t.RequestId = SMGP_ACTIVE_TEST
	t.SequenceId = seq
	return nil
}

func (t *ActiveTest) ToResponse(_ uint32) codec.Pdu {
	resp := ActiveTestRsp{}
	resp.PacketLength = t.PacketLength
	resp.RequestId = SMGP_ACTIVE_TEST_RESP
	resp.SequenceId = t.SequenceId
	return &resp
}

func (t *ActiveTest) Log() []log.Field {
	return (*MessageHeader)(t).Log()
}

func (t *ActiveTest) String() string {
	return (*MessageHeader)(t).String()
}

func (r *ActiveTestRsp) Encode() []byte {
	return (*MessageHeader)(r).Encode()
}

func (r *ActiveTestRsp) Decode(seq uint32, _ []byte) error {
	r.PacketLength = codec.HeadLen
	r.RequestId = SMGP_ACTIVE_TEST_RESP
	r.SequenceId = seq
	return nil
}
func (r *ActiveTestRsp) Log() []log.Field {
	return (*MessageHeader)(r).Log()
}

func (r *ActiveTestRsp) String() string {
	return (*MessageHeader)(r).String()
}
