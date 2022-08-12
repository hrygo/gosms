package sgip

import (
	"encoding/binary"

	"github.com/hrygo/log"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/utils"
)

const BindPkgLen = 61
const BindRspPkgLen = 21

// Bind 登录报文结构体【61 bytes】
type Bind struct {
	MessageHeader        // 【20 bytes】报文头，不含序号
	LoginType     byte   // 【1 bytes 】登录类型。 1:SP 向 SMG 建立的连接，用于发送命令 2:SMG 向 SP 建立的连接，用于发送命令
	LoginName     string // 【16 bytes】服务器端给客户端分配的登录名
	LoginPassword string // 【16 bytes】服务器端和 Login Name 对应的密码
	Reserve       string // 【8 bytes 】保留字段
}

func NewBind(ac *codec.AuthConf, loginType byte) *Bind {
	b := &Bind{}
	b.CommandId = SGIP_BIND
	b.PacketLength = BindPkgLen
	b.SequenceNumber = Sequencer.NextVal()
	b.LoginType = loginType
	b.LoginName = ac.ClientId
	b.LoginPassword = ac.SharedSecret
	return b
}

func (b *Bind) Encode() []byte {
	frame := b.MessageHeader.Encode()
	if len(frame) == BindPkgLen && b.PacketLength == BindPkgLen {
		index := 20
		frame[index] = b.LoginType
		index++
		index = utils.CopyStr(frame, b.LoginName, index, 16)
		index = utils.CopyStr(frame, b.LoginPassword, index, 16)
		index += 16
	}
	return frame
}

func (b *Bind) Decode(cid uint32, frame []byte) error {
	b.PacketLength = codec.HeadLen + uint32(len(frame))
	b.CommandId = SGIP_BIND
	b.SequenceNumber = make([]uint32, 3)
	b.SequenceNumber[0] = cid
	index := 0
	b.SequenceNumber[1] = binary.BigEndian.Uint32(frame[index : index+4])
	index += 4
	b.SequenceNumber[2] = binary.BigEndian.Uint32(frame[index : index+4])
	index += 4
	b.LoginType = frame[index]
	index++
	b.LoginName = utils.TrimStr(frame[index : index+16])
	index += 16
	b.LoginPassword = utils.TrimStr(frame[index : index+16])
	index += 16
	return nil
}

func (b *Bind) Check(ac *codec.AuthConf) Status {
	if ac.LoginName == b.LoginName && ac.SharedSecret == b.LoginPassword {
		return 0
	} else {
		return 1
	}
}

func (b *Bind) ToResponse(code uint32) codec.Pdu {
	rsp := &BindRsp{}
	rsp.PacketLength = 21
	rsp.CommandId = SGIP_BIND_RESP
	rsp.SequenceNumber = b.SequenceNumber
	rsp.Status = Status(code)
	return rsp
}

func (b *Bind) Log() []log.Field {
	ls := b.MessageHeader.Log()
	return append(ls,
		log.Int8("loginType", int8(b.LoginType)),
		log.String("loginName", b.LoginName),
		log.String("loginPassword", "******"),
	)
}

type BindRsp struct {
	MessageHeader
	Status Status
}

func (r *BindRsp) Encode() []byte {
	frame := r.MessageHeader.Encode()
	if len(frame) == BindRspPkgLen && r.PacketLength == BindRspPkgLen {
		frame[20] = byte(r.Status)
	}
	return frame
}

func (r *BindRsp) Decode(cid uint32, frame []byte) error {
	r.PacketLength = codec.HeadLen + uint32(len(frame))
	r.CommandId = SGIP_BIND_RESP
	r.SequenceNumber = make([]uint32, 3)
	r.SequenceNumber[0] = cid
	index := 0
	r.SequenceNumber[1] = binary.BigEndian.Uint32(frame[index : index+4])
	index += 4
	r.SequenceNumber[2] = binary.BigEndian.Uint32(frame[index : index+4])
	index += 4
	r.Status = Status(frame[index])
	return nil
}

func (r *BindRsp) Log() []log.Field {
	ls := r.MessageHeader.Log()
	return append(ls, log.String("status", r.Status.String()))
}
