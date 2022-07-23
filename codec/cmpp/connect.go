package cmpp

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"

	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/utils"
)

// Packet length const for cmpp connect request and response packets.
const (
	ConnReqPktLen   uint32 = 4 + 4 + 4 + 6 + 16 + 1 + 4 // 39d, 0x27
	ConnRspPktLenV2 uint32 = 4 + 4 + 4 + 1 + 16 + 1     // 30d, 0x1e
	ConnRspPktLenV3 uint32 = 4 + 4 + 4 + 4 + 16 + 1     // 33d, 0x21
)

// Errors for connect resp status.
var (
	ErrnoConnInvalidStruct  uint8 = 1
	ErrnoConnInvalidSrcAddr uint8 = 2
	ErrnoConnAuthFailed     uint8 = 3
	ErrnoConnVerTooHigh     uint8 = 4
	ErrnoConnOthers         uint8 = 5

	ConnRspStatusErrMap = map[uint8]error{
		ErrnoConnInvalidStruct:  errConnInvalidStruct,
		ErrnoConnInvalidSrcAddr: errConnInvalidSrcAddr,
		ErrnoConnAuthFailed:     errConnAuthFailed,
		ErrnoConnVerTooHigh:     errConnVerTooHigh,
		ErrnoConnOthers:         errConnOthers,
	}

	errConnInvalidStruct  = errors.New("connect response status: invalid protocol structure")
	errConnInvalidSrcAddr = errors.New("connect response status: invalid source address")
	errConnAuthFailed     = errors.New("connect response status: auth failed")
	errConnVerTooHigh     = errors.New("connect response status: protocol version is too high")
	errConnOthers         = errors.New("connect response status: other errors")
)

// ConnReqPkt represents a Cmpp2 or Cmpp3 connect request packet.
//
// when used in client side(pack), you should initialize it with
// correct SourceAddr(SrcAddr), Secret and Version.
//
// when used in server side(unpack), nothing needed to be initialized.
// unpack will fill the SourceAddr(SrcAddr), AuthSrc, Version, Timestamp
// and SeqId
//
type ConnReqPkt struct {
	SrcAddr   string
	AuthSrc   []byte
	Version   Version
	Timestamp uint32
	Secret    string
	SeqId     uint32
}

// ConnRspPktV2 represents a Cmpp2 connect response packet.
//
// when used in server side(pack), you should initialize it with
// correct Status, AuthSrc, Secret and Version.
//
// when used in client side(unpack), nothing needed to be initialized.
// unpack will fill the Status, AuthImsg, Version and SeqId
//
type ConnRspPktV2 struct {
	Status   uint8
	AuthIsmg []byte
	Version  Version
	Secret   string
	AuthSrc  []byte
	SeqId    uint32
}

// ConnRspPktV3 represents a Cmpp3 connect response packet.
//
// when used in server side(pack), you should initialize it with
// correct Status, AuthSrc, Secret and Version.
//
// when used in client side(unpack), nothing needed to be initialized.
// unpack will fill the Status, AuthImsg, Version and SeqId
//
type ConnRspPktV3 struct {
	Status   uint32
	AuthIsmg []byte
	Version  Version
	Secret   string
	AuthSrc  []byte
	SeqId    uint32
}

// Pack packs the ConnReqPkt to bytes stream for client side.
// Before calling Pack, you should initialize a ConnReqPkt variable
// with correct SourceAddr(SrcAddr), Secret and Version.
func (p *ConnReqPkt) Pack(seqId uint32) ([]byte, error) {
	var w = codec.NewPacketWriter(ConnReqPktLen)

	// Pack header
	w.WriteInt(binary.BigEndian, ConnReqPktLen)
	w.WriteInt(binary.BigEndian, CMPP_CONNECT)
	w.WriteInt(binary.BigEndian, seqId)
	p.SeqId = seqId

	var ts string
	if p.Timestamp == 0 {
		ts, p.Timestamp = utils.Now() // default: current time.
	} else {
		ts = utils.TimeStamp2Str(p.Timestamp)
	}

	// Pack body
	srcAddr := utils.OctetString(p.SrcAddr, 6)
	w.WriteString(srcAddr)

	md5str := md5.Sum(bytes.Join([][]byte{
		[]byte(srcAddr),
		make([]byte, 9),
		[]byte(p.Secret),
		[]byte(ts),
	}, nil))
	p.AuthSrc = md5str[:]

	w.WriteBytes(p.AuthSrc)
	w.WriteInt(binary.BigEndian, p.Version)
	w.WriteInt(binary.BigEndian, p.Timestamp)

	return w.Bytes()
}

// Unpack the binary byte stream to a ConnReqPkt variable.
// Usually it is used in server side. After unpack, you will get SeqId, SourceAddr,
// AuthenticatorSource, Version and Timestamp.
func (p *ConnReqPkt) Unpack(data []byte) error {
	var r = codec.NewPacketReader(data)

	// Sequence Id
	r.ReadInt(binary.BigEndian, &p.SeqId)

	// Body: Source_Addr
	var sa = make([]byte, 6)
	r.ReadBytes(sa)
	p.SrcAddr = string(sa)

	// Body: AuthSrc
	var as = make([]byte, 16)
	r.ReadBytes(as)
	p.AuthSrc = as

	// Body: Version
	r.ReadInt(binary.BigEndian, &p.Version)
	// Body: timestamp
	r.ReadInt(binary.BigEndian, &p.Timestamp)

	return r.Error()
}

// Pack packs the ConnRspPktV2 to bytes stream for server side.
// Before calling Pack, you should initialize a ConnRspPktV2 variable
// with correct Status,AuthenticatorSource, Secret and Version.
func (p *ConnRspPktV2) Pack(seqId uint32) ([]byte, error) {
	var w = codec.NewPacketWriter(ConnRspPktLenV2)

	// pack header
	w.WriteInt(binary.BigEndian, ConnRspPktLenV2)
	w.WriteInt(binary.BigEndian, CMPP_CONNECT_RESP)
	w.WriteInt(binary.BigEndian, seqId)
	p.SeqId = seqId

	// pack body
	w.WriteInt(binary.BigEndian, p.Status)

	md5str := md5.Sum(bytes.Join([][]byte{
		{p.Status},
		[]byte(p.AuthSrc),
		[]byte(p.Secret),
	}, nil))
	p.AuthIsmg = md5str[:]
	w.WriteBytes(p.AuthIsmg)

	w.WriteInt(binary.BigEndian, p.Version)

	return w.Bytes()
}

// Unpack the binary byte stream to a ConnRspPktV2 variable.
// Usually it is used in client side. After unpack, you will get SeqId, Status,
// AuthenticatorIsmg, and Version.
// Parameter data contains seqId in header and the whole packet body.
func (p *ConnRspPktV2) Unpack(data []byte) error {
	var r = codec.NewPacketReader(data)

	// Sequence Id
	r.ReadInt(binary.BigEndian, &p.SeqId)

	// Body: Status
	r.ReadInt(binary.BigEndian, &p.Status)

	// Body: AuthenticatorISMG
	var s = make([]byte, 16)
	r.ReadBytes(s)
	p.AuthIsmg = s

	// Body: Version
	r.ReadInt(binary.BigEndian, &p.Version)
	return r.Error()
}

// Pack packs the ConnRspPktV3 to bytes stream for server side.
// Before calling Pack, you should initialize a ConnRspPktV3 variable
// with correct Status,AuthenticatorSource, Secret and Version.
func (p *ConnRspPktV3) Pack(seqId uint32) ([]byte, error) {
	var w = codec.NewPacketWriter(ConnRspPktLenV3)

	// pack header
	w.WriteInt(binary.BigEndian, ConnRspPktLenV3)
	w.WriteInt(binary.BigEndian, CMPP_CONNECT_RESP)
	w.WriteInt(binary.BigEndian, seqId)
	p.SeqId = seqId

	// pack body
	w.WriteInt(binary.BigEndian, p.Status)

	var statusBuf = new(bytes.Buffer)
	err := binary.Write(statusBuf, binary.BigEndian, p.Status)
	if err != nil {
		return nil, err
	}

	md5str := md5.Sum(bytes.Join([][]byte{
		statusBuf.Bytes(),
		[]byte(p.AuthSrc),
		[]byte(p.Secret),
	}, nil))
	p.AuthIsmg = md5str[:]
	w.WriteBytes(p.AuthIsmg)

	w.WriteInt(binary.BigEndian, p.Version)

	return w.Bytes()
}

// Unpack the binary byte stream to a ConnRspPktV3 variable.
// Usually it is used in client side. After unpack, you will get SeqId, Status,
// AuthenticatorIsmg, and Version.
// Parameter data contains seqId in header and the whole packet body.
func (p *ConnRspPktV3) Unpack(data []byte) error {
	var r = codec.NewPacketReader(data)

	// Sequence Id
	r.ReadInt(binary.BigEndian, &p.SeqId)

	// Body: Status
	r.ReadInt(binary.BigEndian, &p.Status)

	// Body: AuthenticatorISMG
	var s = make([]byte, 16)
	r.ReadBytes(s)
	p.AuthIsmg = s

	// Body: Version
	r.ReadInt(binary.BigEndian, &p.Version)
	return r.Error()
}
