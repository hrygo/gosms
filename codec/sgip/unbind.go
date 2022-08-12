package sgip

import (
	"encoding/binary"

	"github.com/hrygo/log"

	"github.com/hrygo/gosms/codec"
)

type Unbind MessageHeader
type UnbindRsp MessageHeader

func NewUnbind() *Unbind {
	ub := &Unbind{PacketLength: codec.HeadLen + 8, CommandId: SGIP_UNBIND, SequenceNumber: Sequencer.NextVal()}
	return ub
}

func (u *Unbind) Encode() []byte {
	return (*MessageHeader)(u).Encode()
}

func (u *Unbind) Decode(cid uint32, frame []byte) error {
	u.PacketLength = codec.HeadLen + 8
	u.CommandId = SGIP_UNBIND
	u.SequenceNumber = make([]uint32, 3)
	u.SequenceNumber[0] = cid
	u.SequenceNumber[1] = binary.BigEndian.Uint32(frame[0:4])
	u.SequenceNumber[2] = binary.BigEndian.Uint32(frame[4:8])
	return nil
}

func (u *Unbind) ToResponse(_ uint32) codec.Pdu {
	resp := &UnbindRsp{}
	resp.PacketLength = u.PacketLength
	resp.CommandId = SGIP_UNBIND_RESP
	resp.SequenceNumber = u.SequenceNumber
	return resp
}

func (u *Unbind) Log() []log.Field {
	return (*MessageHeader)(u).Log()
}

func (u *Unbind) String() string {
	return (*MessageHeader)(u).String()
}

func (r *UnbindRsp) Encode() []byte {
	return (*MessageHeader)(r).Encode()
}

func (r *UnbindRsp) Decode(cid uint32, frame []byte) error {
	r.PacketLength = codec.HeadLen + 8
	r.CommandId = SGIP_UNBIND_RESP
	r.SequenceNumber = make([]uint32, 3)
	r.SequenceNumber[0] = cid
	r.SequenceNumber[1] = binary.BigEndian.Uint32(frame[0:4])
	r.SequenceNumber[2] = binary.BigEndian.Uint32(frame[4:8])
	return nil
}
func (r *UnbindRsp) Log() []log.Field {
	return (*MessageHeader)(r).Log()
}
