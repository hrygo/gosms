package cmpp

import (
	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/codec"
)

type Terminate MessageHeader

type TerminateRsp MessageHeader

func NewTerminate(seq uint32) *Terminate {
	return &Terminate{TotalLength: HeadLen, CommandId: CMPP_TERMINATE, SequenceId: seq}
}

func (t *Terminate) Encode() []byte {
	return (*MessageHeader)(t).Encode()
}

func (t *Terminate) Decode(seq uint32, _ []byte) error {
	t.TotalLength = HeadLen
	t.CommandId = CMPP_TERMINATE
	t.SequenceId = seq
	return nil
}

func (t *Terminate) ToResponse(_ uint32) codec.Pdu {
	rsp := &TerminateRsp{}
	rsp.TotalLength = HeadLen
	rsp.CommandId = CMPP_TERMINATE_RESP
	rsp.SequenceId = t.SequenceId
	return rsp
}

func (t *Terminate) Log() []log.Field {
	return (*MessageHeader)(t).Log()
}

func (r *TerminateRsp) Encode() []byte {
	return (*MessageHeader)(r).Encode()
}

func (r *TerminateRsp) Decode(seqId uint32, _ []byte) error {
	r.TotalLength = HeadLen
	r.CommandId = CMPP_TERMINATE_RESP
	r.SequenceId = seqId
	return nil
}

func (r *TerminateRsp) Log() []log.Field {
	return (*MessageHeader)(r).Log()
}
