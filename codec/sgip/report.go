package sgip

import (
	"encoding/binary"
	"fmt"

	"github.com/hrygo/log"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/utils"
)

type Report struct {
	MessageHeader
	MtSequence []uint32 //  该命令所涉及的 Submit 或 deliver 命令的序列号
	ReportType byte     //  Report 命令类型 0:对先前一条 Submit 命令的状态报告 1:对先前一条前转 Deliver 命令的状态报告
	UserNumber string   //  接收短消息的手机号，手机号码前加“86”国别标 志
	State      Status   //  该命令所涉及的短消息的当前执行状态 0:发送成功 1:等待发送 2:发送失败
	ErrorCode  byte     //  当 State=2 时为错误码值，否则为 0
	Reserve    string   //  保留，扩展用【 8 bytes 】
}

const DlvPackLen = 64

func NewReport(phone string, mtSequence []uint32, status Status, code byte) codec.RequestPdu {
	dlv := &Report{}
	dlv.PacketLength = DlvPackLen
	dlv.CommandId = SGIP_DELIVER
	dlv.SequenceNumber = Sequencer.NextVal()
	dlv.MtSequence = mtSequence
	dlv.UserNumber = phone
	dlv.State = status
	dlv.ErrorCode = code
	return dlv
}

func (r *Report) Encode() []byte {
	frame := r.MessageHeader.Encode()
	index := 20
	for _, seq := range r.MtSequence {
		binary.BigEndian.PutUint32(frame[index:], seq)
		index += 4
	}
	frame[index] = r.ReportType
	index++
	copy(frame[index:], r.UserNumber)
	index += 21
	frame[index] = byte(r.State)
	index++
	frame[index] = r.ErrorCode
	index++
	copy(frame[index:], r.Reserve)
	return frame
}

func (r *Report) Decode(cid uint32, frame []byte) error {
	r.PacketLength = codec.HeadLen + uint32(len(frame))
	r.CommandId = SGIP_REPORT
	r.SequenceNumber = make([]uint32, 3)
	r.SequenceNumber[0] = cid
	index := 0
	r.SequenceNumber[1] = binary.BigEndian.Uint32(frame[index:])
	index += 4
	r.SequenceNumber[2] = binary.BigEndian.Uint32(frame[index:])
	index += 4
	r.MtSequence = make([]uint32, 3)
	r.SequenceNumber[0] = binary.BigEndian.Uint32(frame[index:])
	index += 4
	r.SequenceNumber[1] = binary.BigEndian.Uint32(frame[index:])
	index += 4
	r.SequenceNumber[2] = binary.BigEndian.Uint32(frame[index:])
	index += 4
	r.ReportType = frame[index]
	index++
	r.UserNumber = utils.TrimStr(frame[index : index+21])
	index += 21
	r.State = Status(frame[index])
	index++
	r.ErrorCode = frame[index]
	return nil
}

func (r *Report) Log() []log.Field {
	ls := r.MessageHeader.Log()
	return append(ls,
		log.Uint64(codec.Seq, uint64(r.MtSequence[1])<<32|uint64(r.MtSequence[2])),
		log.String(codec.Seq+"_12", fmt.Sprintf("%010d%010d%08x", r.MtSequence[0], r.MtSequence[1], r.MtSequence[2])),
		log.Uint8("reportType", r.ReportType),
		log.String("userNumber", r.UserNumber),
		log.String("status", r.State.String()),
		log.Uint8("errorCode", r.ErrorCode),
	)
}

type ReportRsp struct {
	MessageHeader
	Status  Status
	Reserve string
}

func (r *Report) ToResponse(code uint32) codec.Pdu {
	rsp := &ReportRsp{}
	rsp.PacketLength = codec.HeadLen + 8 + 1 + 8
	rsp.CommandId = SGIP_REPORT_RESP
	rsp.SequenceNumber = r.SequenceNumber
	rsp.Status = Status(code)
	rsp.Reserve = ""
	return rsp
}

func (r *ReportRsp) Decode(cid uint32, frame []byte) error {
	r.PacketLength = codec.HeadLen + uint32(len(frame))
	r.CommandId = SGIP_REPORT_RESP
	r.SequenceNumber = make([]uint32, 3)
	r.SequenceNumber[0] = cid
	r.SequenceNumber[1] = binary.BigEndian.Uint32(frame[0:4])
	r.SequenceNumber[2] = binary.BigEndian.Uint32(frame[4:8])
	r.Status = Status(frame[8])
	r.Reserve = ""
	return nil
}

func (r *ReportRsp) Encode() []byte {
	frame := r.MessageHeader.Encode()
	frame[20] = byte(r.Status)
	return frame
}

func (r *ReportRsp) Log() []log.Field {
	ls := r.MessageHeader.Log()
	return append(ls, log.String("status", r.Status.String()))
}
