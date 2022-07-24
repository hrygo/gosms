package cmpp

import (
	"encoding/binary"

	"github.com/hrygo/gosmsn/codec"
)

// Packet length const for cmpp active test request and response packets.
const (
	ActiveTestReqPktLen uint32 = 12     // 12d, 0xc
	ActiveTestRspPktLen uint32 = 12 + 1 // 13d, 0xd
)

type ActiveTestReqPkt struct {
	// session info
	SeqId uint32
}
type ActiveTestRspPkt struct {
	Reserved uint8
	// session info
	SeqId uint32
}

// Pack packs the CmppActiveTestReqPkt to bytes stream for client side.
func (p *ActiveTestReqPkt) Pack(seqId uint32) ([]byte, error) {
	var pktLen = ActiveTestReqPktLen

	var w = codec.NewPacketWriter(pktLen)

	// Pack header
	w.WriteInt(binary.BigEndian, pktLen)
	w.WriteInt(binary.BigEndian, CMPP_ACTIVE_TEST)
	w.WriteInt(binary.BigEndian, seqId)
	p.SeqId = seqId

	return w.Bytes()
}

// Unpack the binary byte stream to a CmppActiveTestReqPkt variable.
// After unpack, you will get all value of fields in
// CmppActiveTestReqPkt struct.
func (p *ActiveTestReqPkt) Unpack(data []byte) error {
	var r = codec.NewPacketReader(data)

	// Sequence Id
	r.ReadInt(binary.BigEndian, &p.SeqId)
	return r.Error()
}

// Pack packs the CmppActiveTestRspPkt to bytes stream for client side.
func (p *ActiveTestRspPkt) Pack(seqId uint32) ([]byte, error) {
	var pktLen = ActiveTestRspPktLen

	var w = codec.NewPacketWriter(pktLen)

	// Pack header
	w.WriteInt(binary.BigEndian, pktLen)
	w.WriteInt(binary.BigEndian, CMPP_ACTIVE_TEST_RESP)
	w.WriteInt(binary.BigEndian, seqId)
	w.WriteByte(p.Reserved)
	p.SeqId = seqId

	return w.Bytes()
}

// Unpack the binary byte stream to a CmppActiveTestRspPkt variable.
// After unpack, you will get all value of fields in
// CmppActiveTestRspPkt struct.
func (p *ActiveTestRspPkt) Unpack(data []byte) error {
	var r = codec.NewPacketReader(data)

	// Sequence Id
	r.ReadInt(binary.BigEndian, &p.SeqId)
	p.Reserved = r.ReadByte()
	return r.Error()
}
