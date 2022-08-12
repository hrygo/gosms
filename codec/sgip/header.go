package sgip

import (
	"encoding/binary"
	"fmt"

	"github.com/hrygo/log"

	"github.com/hrygo/gosms/codec"
)

// MessageHeader 报文头【20 bytes】
type MessageHeader struct {
	PacketLength   uint32    // 【4 bytes】数据包总长度
	CommandId      CommandId // 【4 bytes】命令类型
	SequenceNumber []uint32  // 【12 bytes】序号
}

func (h *MessageHeader) Encode() []byte {
	if h.PacketLength < codec.HeadLen+8 {
		h.PacketLength = codec.HeadLen + 8
	}
	frame := make([]byte, h.PacketLength)
	binary.BigEndian.PutUint32(frame[0:4], h.PacketLength)
	binary.BigEndian.PutUint32(frame[4:8], uint32(h.CommandId))
	binary.BigEndian.PutUint32(frame[8:12], h.SequenceNumber[0])
	binary.BigEndian.PutUint32(frame[12:16], h.SequenceNumber[1])
	binary.BigEndian.PutUint32(frame[16:20], h.SequenceNumber[2])
	return frame
}

func (h *MessageHeader) Decode(frame []byte) error {
	h.PacketLength = binary.BigEndian.Uint32(frame[0:4])
	h.CommandId = CommandId(binary.BigEndian.Uint32(frame[4:8]))
	h.SequenceNumber = make([]uint32, 3)
	h.SequenceNumber[0] = binary.BigEndian.Uint32(frame[8:12])
	h.SequenceNumber[1] = binary.BigEndian.Uint32(frame[12:16])
	h.SequenceNumber[2] = binary.BigEndian.Uint32(frame[16:20])
	return nil
}

func (h *MessageHeader) Log() []log.Field {
	ls := make([]log.Field, 0, 16)
	ls = append(ls, log.Uint32(codec.Pkl, h.PacketLength),
		log.String(codec.Cmd, h.CommandId.String()),
		log.Uint64(codec.Seq, uint64(h.SequenceNumber[1])<<32|uint64(h.SequenceNumber[2])),
		log.String(codec.Seq+"_12", fmt.Sprintf("%010d%010d%08x", h.SequenceNumber[0], h.SequenceNumber[1], h.SequenceNumber[2])))
	return ls
}

func (h *MessageHeader) String() string {
	return fmt.Sprintf("{%s: %d, %s: %s, %s: %s}",
		codec.Pkl, h.PacketLength,
		codec.Cmd, h.CommandId.String(),
		codec.Seq, fmt.Sprintf("%010d%010d%08x", h.SequenceNumber[0], h.SequenceNumber[1], h.SequenceNumber[2]))
}
