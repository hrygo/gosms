package smgp

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/auth"
	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/utils"
)

type Deliver struct {
	MessageHeader
	msgId      []byte         // 【10字节】短消息流水号
	isReport   byte           // 【1字节】是否为状态报告
	msgFormat  byte           // 【1字节】短消息格式
	recvTime   string         // 【14字节】短消息定时发送时间
	srcTermID  string         // 【21字节】短信息发送方号码
	destTermID string         // 【21】短消息接收号码
	msgLength  byte           // 【1字节】短消息长度
	msgContent string         // 【MsgLength字节】短消息内容
	msgBytes   []byte         // 消息内容按照Msg_Fmt编码后的数据
	report     *Report        // 状态报告
	reserve    string         // 【8字节】保留
	tlvList    *utils.TlvList // 【TLV】可选项参数

	// 协议版本,不是报文内容，但在调用encode方法前需要设置此值
	Version Version
}

type DeliverRsp struct {
	MessageHeader
	msgId  []byte // 【10字节】短消息流水号
	status Status

	// 协议版本,不是报文内容，但在调用encode方法前需要设置此值
	Version Version
}

func NewDeliver(cli *auth.Client, phone string, destNo string, txt string, seq uint32) codec.RequestPdu {
	baseLen := uint32(89)
	dlv := &Deliver{Version: Version(cli.Version)}
	dlv.RequestId = SMGP_DELIVER
	dlv.SequenceId = seq
	dlv.msgId = codec.BcdSeq.NextVal()
	dlv.isReport = 0
	dlv.msgFormat = 15
	dlv.recvTime = time.Now().Format("20060102150405")
	dlv.srcTermID = phone
	dlv.destTermID = cli.SmsDisplayNo + destNo
	// 上行最长70字符
	subTxt := txt
	rs := []rune(txt)
	if len(rs) > 70 {
		rs = rs[:70]
		subTxt = string(rs)
	}
	gbs, _ := GbEncoder.String(subTxt)
	msg := []byte(gbs)
	dlv.msgBytes = msg
	dlv.msgLength = byte(len(msg))
	dlv.msgContent = subTxt
	dlv.PacketLength = baseLen + uint32(dlv.msgLength)
	return dlv
}

func NewDeliveryReport(cli *auth.Client, mt *Submit, seq uint32, msgId []byte) *Deliver {
	baseLen := uint32(89)
	dlv := &Deliver{Version: Version(cli.Version)}
	dlv.RequestId = SMGP_DELIVER
	dlv.SequenceId = seq
	dlv.msgId = codec.BcdSeq.NextVal()
	dlv.report = NewReport(msgId)
	dlv.msgLength = 115
	dlv.isReport = 1
	dlv.msgFormat = 0
	dlv.recvTime = time.Now().Format("20060102150405")
	dlv.srcTermID = mt.destTermID[0]
	dlv.destTermID = mt.srcTermID
	dlv.PacketLength = baseLen + uint32(RptLen)
	return dlv
}

func (d *Deliver) Encode() []byte {
	frame := d.MessageHeader.Encode()
	index := 12
	copy(frame[index:index+10], d.msgId)
	index += 10
	index = utils.CopyByte(frame, d.isReport, index)
	index = utils.CopyByte(frame, d.msgFormat, index)
	index = utils.CopyStr(frame, d.recvTime, index, 14)
	index = utils.CopyStr(frame, d.srcTermID, index, 21)
	index = utils.CopyStr(frame, d.destTermID, index, 21)
	index = utils.CopyByte(frame, d.msgLength, index)
	if d.IsReport() && d.report != nil {
		rts := d.report.Encode()
		copy(frame[index:index+RptLen], rts)
		index += RptLen
	} else {
		copy(frame[index:index+int(d.msgLength)], d.msgBytes)
		index += int(d.msgLength)
	}
	index = utils.CopyStr(frame, d.reserve, index, 8)
	return frame
}

func (d *Deliver) Decode(seq uint32, frame []byte) error {
	d.PacketLength = codec.HeadLen + uint32(len(frame))
	d.RequestId = SMGP_DELIVER
	d.SequenceId = seq
	var index int
	d.msgId = frame[index : index+10]
	index += 10
	d.isReport = frame[index]
	index += 1
	d.msgFormat = frame[index]
	index += 1
	d.recvTime = utils.TrimStr(frame[index : index+14])
	index += 14
	d.srcTermID = utils.TrimStr(frame[index : index+21])
	index += 21
	d.destTermID = utils.TrimStr(frame[index : index+21])
	index += 21
	d.msgLength = frame[index]
	index += 1
	if d.IsReport() {
		d.report = &Report{}
		err := d.report.Decode(frame[index : index+RptLen])
		if err != nil {
			return err
		}
	} else {
		bytes, err := GbDecoder.Bytes(frame[index : index+int(d.msgLength)])
		if err != nil {
			return err
		}
		d.msgContent = string(bytes)
	}
	// 后续字节不解析了
	return nil
}

func (d *Deliver) ToResponse(code uint32) codec.Pdu {
	resp := &DeliverRsp{Version: d.Version}
	resp.RequestId = SMGP_DELIVER_RESP
	resp.PacketLength = 26
	resp.SequenceId = d.SequenceId
	resp.status = Status(code)
	resp.msgId = codec.BcdSeq.NextVal()
	return resp
}

func (d *Deliver) String() string {
	content := ""
	if d.IsReport() {
		content = d.report.String()
	} else {
		content = strings.ReplaceAll(d.msgContent, "\n", " ")
	}
	return fmt.Sprintf("{ header: %v, msgId: %x, isReport: %v, msgFormat: %v, recvTime: %v,"+
		" SrcTermID: %v, destTermID: %v, msgLength: %v, "+
		"msgContent: \"%s\", reserve: %v, tlv: %v }",
		d.MessageHeader, d.msgId, d.isReport, d.msgFormat, d.recvTime,
		d.srcTermID, d.destTermID, d.msgLength,
		content, d.reserve, d.tlvList,
	)
}

func (d *Deliver) Log() []log.Field {
	ls := d.MessageHeader.Log()
	ls = append(ls,
		log.String("version", hex.EncodeToString([]byte{byte(d.Version)})),
		log.String("msgId", hex.EncodeToString(d.msgId)),
		log.Uint8("isReport", d.isReport),
		log.Uint8("msgFmt", d.msgFormat),
		log.String("recvTime", d.recvTime),
		log.String("srcTermID", d.srcTermID),
		log.String("destTermID", d.destTermID),
		log.Uint8("msgLength", d.msgLength),
		log.String("reserve", d.reserve),
		log.String("tlv", d.tlvList.String()),
	)
	var csl log.Field
	var bs = []byte(d.msgContent)
	var l = len(bs)
	if l > 6 {
		l = 6
	}
	if d.isReport == 1 {
		csl = log.String("msgContent", d.report.String())
	} else {
		csl = log.String("msgContent", hex.EncodeToString(bs[:l])+"...")
	}
	return append(ls, csl)
}

func (d *Deliver) IsReport() bool {
	return d.isReport == 1
}

func (r *DeliverRsp) Encode() []byte {
	frame := r.MessageHeader.Encode()
	index := 12
	copy(frame[index:index+10], r.msgId)
	index += 10
	binary.BigEndian.PutUint32(frame[index:index+4], uint32(r.status))
	return frame
}

func (r *DeliverRsp) Decode(seq uint32, frame []byte) error {
	r.PacketLength = codec.HeadLen + uint32(len(frame))
	r.RequestId = SMGP_DELIVER_RESP
	r.SequenceId = seq
	r.msgId = make([]byte, 10)
	copy(r.msgId, frame[0:10])
	r.status = Status(binary.BigEndian.Uint32(frame[10:14]))
	return nil
}

func (r *DeliverRsp) String() string {
	return fmt.Sprintf("{ header: %s, msgId: %x, status: \"%s\" }", &r.MessageHeader, r.msgId, r.status)
}

func (r *DeliverRsp) Log() (rt []log.Field) {
	rt = r.MessageHeader.Log()
	rt = append(rt,
		log.String("version", hex.EncodeToString([]byte{byte(r.Version)})),
		log.String("msgId", hex.EncodeToString(r.msgId)),
		log.String("status", r.status.String()),
	)
	return
}
func (r *DeliverRsp) MsgId() string {
	return fmt.Sprintf("%x", r.msgId)
}

func (d *Deliver) MsgId() []byte {
	return d.msgId
}

func (d *Deliver) MsgFormat() byte {
	return d.msgFormat
}

func (d *Deliver) RecvTime() string {
	return d.recvTime
}

func (d *Deliver) SrcTermID() string {
	return d.srcTermID
}

func (d *Deliver) DestTermID() string {
	return d.destTermID
}

func (d *Deliver) MsgLength() byte {
	return d.msgLength
}

func (d *Deliver) MsgContent() string {
	return d.msgContent
}

func (d *Deliver) MsgBytes() []byte {
	return d.msgBytes
}

func (d *Deliver) Report() *Report {
	return d.report
}

func (d *Deliver) Reserve() string {
	return d.reserve
}

func (d *Deliver) TlvList() *utils.TlvList {
	return d.tlvList
}

func (r *DeliverRsp) Status() Status {
	return r.status
}
