package cmpp

import (
	"encoding/binary"

	"github.com/hrygo/log"

	"github.com/hrygo/gosms/codec"
)

type MessageHeader struct {
	TotalLength uint32
	CommandId   CommandId
	SequenceId  uint32
}

func (header *MessageHeader) Encode() []byte {
	if header.TotalLength < codec.HeadLen {
		header.TotalLength = codec.HeadLen
	}
	frame := make([]byte, header.TotalLength)
	binary.BigEndian.PutUint32(frame[0:4], header.TotalLength)
	binary.BigEndian.PutUint32(frame[4:8], uint32(header.CommandId))
	binary.BigEndian.PutUint32(frame[8:12], header.SequenceId)
	return frame
}

func (header *MessageHeader) Decode(frame []byte) error {
	header.TotalLength = binary.BigEndian.Uint32(frame[0:4])
	header.CommandId = CommandId(binary.BigEndian.Uint32(frame[4:8]))
	header.SequenceId = binary.BigEndian.Uint32(frame[8:12])
	return nil
}

func (header *MessageHeader) Log() []log.Field {
	ls := make([]log.Field, 0, 16)
	ls = append(ls, log.Uint32(codec.Pkl, header.TotalLength),
		log.String(codec.Cmd, header.CommandId.String()),
		log.Uint32(codec.Seq, header.SequenceId))
	return ls
}
