package cmpp

import (
	"github.com/hrygo/log"

	"github.com/hrygo/gosms/codec"
)

type ActiveTest MessageHeader

type ActiveTestRsp MessageHeader

func NewActiveTest(seq uint32) *ActiveTest {
	return &ActiveTest{TotalLength: codec.HeadLen, CommandId: CMPP_ACTIVE_TEST, SequenceId: seq}
}

func (at *ActiveTest) Encode() []byte {
	return (*MessageHeader)(at).Encode()
}

func (at *ActiveTest) Decode(seq uint32, _ []byte) error {
	at.TotalLength = 12
	at.CommandId = CMPP_ACTIVE_TEST
	at.SequenceId = seq
	return nil
}

func (at *ActiveTest) ToResponse(_ uint32) codec.Pdu {
	rsp := &ActiveTestRsp{}
	rsp.TotalLength = 12 + 1
	rsp.CommandId = CMPP_ACTIVE_TEST_RESP
	rsp.SequenceId = at.SequenceId
	return rsp
}

func (at *ActiveTest) Log() []log.Field {
	return (*MessageHeader)(at).Log()
}

func (at *ActiveTestRsp) Encode() []byte {
	ls := (*MessageHeader)(at).Encode()
	return append(ls, 0)
}

func (at *ActiveTestRsp) Decode(seqId uint32, _ []byte) error {
	at.TotalLength = codec.HeadLen + 1
	at.CommandId = CMPP_ACTIVE_TEST_RESP
	at.SequenceId = seqId
	return nil
}

func (at *ActiveTestRsp) Log() []log.Field {
	return (*MessageHeader)(at).Log()
}
