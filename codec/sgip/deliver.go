package sgip

import (
	"encoding/binary"
	"encoding/hex"
	"strings"

	"github.com/hrygo/log"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/utils"
)

type Deliver struct {
	MessageHeader
	UserNumber     string //  接收该短消息的手机号，该字段重复UserCount指定的次数，手机号码前加“86”国别标志【 21 bytes 】
	SPNumber       string //  SP的接入号码【 21 bytes 】
	TpPid          byte   //  GSM协议类型。详细解释请参考GSM03.40中的9.2.3.9 【 1  bytes 】
	TpUdhi         byte   //  GSM协议类型。详细解释请参考GSM03.40中的9.2.3.9 【 1  bytes 】
	MessageCoding  byte   //  短消息的编码格式。 【 1  bytes 】
	MessageLength  uint32 //  短消息的长度【 4 bytes 】
	MessageContent []byte //  编码后消息内容
	Reserve        string //  保留，扩展用【 8 bytes 】
}

const MoBaseLen = 77

func NewDeliver(ac *codec.AuthConf, phone, content, destNo string) codec.RequestPdu {
	dlv := &Deliver{}
	dlv.PacketLength = MoBaseLen
	dlv.CommandId = SGIP_DELIVER
	dlv.SequenceNumber = Sequencer.NextVal()
	dlv.UserNumber = phone
	if !strings.HasPrefix(destNo, ac.SmsDisplayNo) {
		destNo = ac.SmsDisplayNo + destNo
	}
	dlv.SPNumber = destNo
	dlv.MessageCoding = utils.MsgFmt(content)

	var bs []byte
	// 上行短信不支持长短信，过长内容会被截取
	if dlv.MessageCoding == 8 {
		bs, _ = utils.Utf8ToUcs2(content)
		if len(bs) > 140 {
			bs = bs[:140]
		}
	} else {
		bs = []byte(content)
		if len(bs) > 160 {
			bs = bs[:160]
		}
	}
	dlv.MessageLength = uint32(len(bs))
	dlv.MessageContent = bs
	dlv.PacketLength = MoBaseLen + dlv.MessageLength
	return dlv
}

func (d *Deliver) Encode() []byte {
	frame := d.MessageHeader.Encode()
	index := 20
	copy(frame[index:], d.UserNumber)
	index += 21
	copy(frame[index:], d.SPNumber)
	index += 21
	frame[index] = d.TpPid
	index++
	frame[index] = d.TpUdhi
	index++
	frame[index] = d.MessageCoding
	index++
	binary.BigEndian.PutUint32(frame[index:], d.MessageLength)
	index += 4
	copy(frame[index:], d.MessageContent)
	index += len(d.MessageContent)
	copy(frame[index:], d.Reserve)
	return frame
}

func (d *Deliver) Decode(cid uint32, frame []byte) error {
	d.PacketLength = codec.HeadLen + uint32(len(frame))
	d.CommandId = SGIP_DELIVER
	d.SequenceNumber = make([]uint32, 3)
	d.SequenceNumber[0] = cid
	index := 0
	d.SequenceNumber[1] = binary.BigEndian.Uint32(frame[index:])
	index += 4
	d.SequenceNumber[2] = binary.BigEndian.Uint32(frame[index:])
	index += 4
	d.UserNumber = utils.TrimStr(frame[index : index+21])
	index += 21
	d.SPNumber = utils.TrimStr(frame[index : index+21])
	index += 21
	d.TpPid = frame[index]
	index++
	d.TpUdhi = frame[index]
	index++
	d.MessageCoding = frame[index]
	index++
	d.MessageLength = binary.BigEndian.Uint32(frame[index:])
	index += 4
	d.MessageContent = frame[index : index+int(d.MessageLength)]
	index += int(d.MessageLength)
	d.Reserve = ""
	return nil
}

func (d *Deliver) Log() []log.Field {
	ls := d.MessageHeader.Log()
	var l = len(d.MessageContent)
	if l > 6 {
		l = 6
	}
	msg := hex.EncodeToString(d.MessageContent[:l]) + "..."
	return append(ls,
		log.String("userNumber", d.UserNumber),
		log.String("spNumber", d.SPNumber),
		log.Uint8("msgFormat", d.MessageCoding),
		log.Uint32("msgLength", d.MessageLength),
		log.String("msgContent", msg),
		log.Uint8("tpPid", d.TpPid),
		log.Uint8("tpUdhi", d.TpUdhi),
	)
}

type DeliverRsp struct {
	MessageHeader
	Status  Status
	Reserve string
}

func (d *Deliver) ToResponse(code uint32) codec.Pdu {
	rsp := &DeliverRsp{}
	rsp.PacketLength = codec.HeadLen + 8 + 1 + 8
	rsp.CommandId = SGIP_DELIVER_RESP
	rsp.SequenceNumber = d.SequenceNumber
	rsp.Status = Status(code)
	rsp.Reserve = ""
	return rsp
}

func (r *DeliverRsp) Decode(cid uint32, frame []byte) error {
	r.PacketLength = codec.HeadLen + uint32(len(frame))
	r.CommandId = SGIP_SUBMIT_RESP
	r.SequenceNumber = make([]uint32, 3)
	r.SequenceNumber[0] = cid
	r.SequenceNumber[1] = binary.BigEndian.Uint32(frame[0:4])
	r.SequenceNumber[2] = binary.BigEndian.Uint32(frame[4:8])
	r.Status = Status(frame[8])
	r.Reserve = ""
	return nil
}

func (r *DeliverRsp) Encode() []byte {
	frame := r.MessageHeader.Encode()
	frame[20] = byte(r.Status)
	return frame
}

func (r *DeliverRsp) Log() []log.Field {
	ls := r.MessageHeader.Log()
	return append(ls, log.String("status", r.Status.String()))
}
