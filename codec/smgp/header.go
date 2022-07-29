package smgp

import (
	"encoding/binary"
	"fmt"

	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/codec"
)

type MessageHeader struct {
	PacketLength uint32
	RequestId    CommandId
	SequenceId   uint32
}

func (h *MessageHeader) Encode() []byte {
	if h.PacketLength < codec.HeadLen {
		h.PacketLength = codec.HeadLen
	}
	frame := make([]byte, h.PacketLength)
	binary.BigEndian.PutUint32(frame[0:4], h.PacketLength)
	binary.BigEndian.PutUint32(frame[4:8], uint32(h.RequestId))
	binary.BigEndian.PutUint32(frame[8:12], h.SequenceId)
	return frame
}

func (h *MessageHeader) Decode(frame []byte) error {
	h.PacketLength = binary.BigEndian.Uint32(frame[0:4])
	h.RequestId = CommandId(binary.BigEndian.Uint32(frame[4:8]))
	h.SequenceId = binary.BigEndian.Uint32(frame[8:12])
	return nil
}

func (h *MessageHeader) Log() []log.Field {
	ls := make([]log.Field, 0, 16)
	ls = append(ls, log.Uint32(codec.Pkl, h.PacketLength),
		log.String(codec.Cmd, h.RequestId.String()),
		log.Uint32(codec.Seq, h.SequenceId))
	return ls
}

func (h *MessageHeader) String() string {
	return fmt.Sprintf("{%s: %d, %s: %s, %s: %d}", codec.Pkl, h.PacketLength, codec.Cmd, h.RequestId.String(), codec.Seq, h.SequenceId)
}
