package cmpp

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/auth"
	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/utils"
)

// Delivery 上行短信或状态报告，不支持长短信
type Delivery struct {
	MessageHeader

	msgId              uint64  // 信息标识
	destId             string  // 目的号码 21
	serviceId          string  // 业务标识，是数字、字母和符号的组合。 10
	tpPid              uint8   // 见Submit
	tpUdhi             uint8   // 见Submit
	msgFmt             uint8   // 见Submit
	srcTerminalId      string  // 源终端MSISDN号码（状态报告时填为CMPP_SUBMIT消息的目的终端号码）
	srcTerminalType    uint8   // 源终端号码类型，0：真实号码；1：伪码
	registeredDelivery uint8   // 是否为状态报告
	msgLength          uint8   // 消息长度
	msgContent         string  // 非状态报告的消息内容
	report             *Report // 状态报告的消息内容
	linkID             string  // 点播业务使用的LinkID，非点播类业务的MT流程不使用该字段

	// 协议版本,不是报文内容，但在调用encode方法前需要设置此值
	Version Version
}

func NewDelivery(cli *auth.Client, phone, msg, dest, serviceId string, seq uint32) codec.RequestPdu {
	dly := &Delivery{Version: Version(cli.Version)}
	dly.CommandId = CMPP_DELIVER
	dly.SequenceId = seq
	dly.srcTerminalId = phone
	dly.srcTerminalType = 0
	setMsgContent(dly, msg)

	if dest != "" {
		dly.destId = dest
	} else {
		dly.destId = cli.SmsDisplayNo
	}
	if serviceId != "" {
		dly.serviceId = serviceId
	} else {
		dly.serviceId = cli.ServiceId
	}
	baseLen := uint32(85)
	if V30.MajorMatch(cli.Version) {
		baseLen = 109
	}
	dly.TotalLength = baseLen + uint32(dly.msgLength)
	return dly
}

// Encode 调用前需设置版本号 Version
func (d *Delivery) Encode() []byte {
	frame := d.MessageHeader.Encode()
	binary.BigEndian.PutUint64(frame[12:20], d.msgId)
	copy(frame[20:41], d.destId)
	copy(frame[41:51], d.serviceId)
	frame[51] = d.tpPid
	frame[52] = d.tpUdhi
	frame[53] = d.msgFmt
	index := 54
	if V30.MajorMatchV(d.Version) {
		copy(frame[index:index+32], d.srcTerminalId)
		index += 32
		frame[index] = d.srcTerminalType
		index++
	} else {
		copy(frame[index:index+21], d.srcTerminalId)
		index += 21
	}
	frame[index] = d.registeredDelivery
	index++
	frame[index] = d.msgLength
	index++
	l := int(d.msgLength)
	if d.registeredDelivery == 1 {
		// 状态报告
		copy(frame[index:index+l], d.report.Encode())
	} else {
		// 上行短信，不支持长短信，固定选用第一片 （New时需处理）
		slices := MsgSlices(d.msgFmt, d.msgContent)
		// 不支持长短信，固定选用第一片
		content := slices[0]
		copy(frame[index:index+l], content)
	}
	index += l
	if V30.MajorMatchV(d.Version) {
		copy(frame[index:index+20], d.linkID)
	}

	return frame
}

func (d *Delivery) Decode(seq uint32, frame []byte) error {
	d.TotalLength = codec.HeadLen + uint32(len(frame))
	d.CommandId = CMPP_DELIVER
	d.SequenceId = seq
	d.msgId = binary.BigEndian.Uint64(frame[0:8])
	d.destId = utils.TrimStr(frame[8:29])
	d.destId = utils.TrimStr(frame[29:39])
	d.tpPid = frame[39]
	d.tpUdhi = frame[40]
	d.msgFmt = frame[41]
	index := 42
	if V30.MajorMatchV(d.Version) {
		d.srcTerminalId = utils.TrimStr(frame[index : index+32])
		index += 32
		d.srcTerminalType = frame[index]
		index++
	} else {
		d.srcTerminalId = utils.TrimStr(frame[index : index+21])
		index += 21
	}
	d.registeredDelivery = frame[index]
	index++
	d.msgLength = frame[index]
	index++
	l := int(d.msgLength)
	if d.registeredDelivery == 1 {
		rpt := &Report{}
		err := rpt.Decode(frame[index : index+l])
		if err != nil {
			return err
		}
		d.report = rpt
	} else {
		d.msgContent = utils.TrimStr(frame[index : index+l])
	}
	index += l
	if V30.MajorMatchV(d.Version) {
		d.linkID = utils.TrimStr(frame[index : index+20])
	}
	return nil
}

func (d *Delivery) ToResponse(code uint32) codec.Pdu {
	dr := &DeliveryRsp{}
	dr.TotalLength = codec.HeadLen + 9
	if V30.MajorMatchV(d.Version) {
		dr.TotalLength = codec.HeadLen + 12
	}
	dr.CommandId = CMPP_DELIVER_RESP
	dr.msgId = d.msgId
	dr.result = DlyResult(code)

	dr.Version = d.Version
	return dr
}

func setMsgContent(dly *Delivery, msg string) {
	dly.msgFmt = MsgFmt(msg)
	var l int
	if dly.msgFmt == 8 {
		l = 2 * len([]rune(msg))
		if l > 140 {
			// 只取前70个字符
			rs := []rune(msg)
			msg = string(rs[:70])
			l = 140
		}
	} else {
		l = len(msg)
		if l > 160 {
			// 只取前160个字符
			msg = msg[:160]
			l = 160
		}
	}
	dly.msgLength = uint8(l)
	dly.msgContent = msg
}

func (d *Delivery) RegisteredDelivery() uint8 {
	return d.registeredDelivery
}

func (d *Delivery) Log() []log.Field {
	ls := d.MessageHeader.Log()
	ls = append(ls,
		log.String("version", hex.EncodeToString([]byte{byte(d.Version)})),
		log.String("msgId", utils.Uint64HexString(d.msgId)),
		log.Uint8("isReport", d.registeredDelivery),
		log.Uint8("msgFmt", d.msgFmt),
		log.Uint8("msgLength", d.msgLength),
		log.String("destId", d.destId),
		log.String("serviceId", d.serviceId),
		log.Uint8("srcTerminalType", d.srcTerminalType),
		log.String("srcTerminalId", d.srcTerminalId),
		log.Uint8("tpPid", d.tpPid),
		log.Uint8("tpUdhi", d.tpUdhi),
		log.String("linkID", d.linkID),
	)
	var csl log.Field
	var bs = []byte(d.msgContent)
	var l = len(bs)
	if l > 6 {
		l = 6
	}
	if d.registeredDelivery == 1 {
		csl = log.String("msgContent", d.report.String())
	} else {
		csl = log.String("msgContent", hex.EncodeToString(bs[:l])+"...")
	}
	return append(ls, csl)
}

type DeliveryRsp struct {
	MessageHeader
	msgId  uint64    // 消息标识,来自CMPP_DELIVERY
	result DlyResult // 结果

	// 协议版本,不是报文内容，ToResponse 会设置此值
	Version Version
}

func (r *DeliveryRsp) Encode() []byte {
	frame := r.MessageHeader.Encode()
	binary.BigEndian.PutUint64(frame[12:20], r.msgId)
	if V30.MajorMatchV(r.Version) {
		binary.BigEndian.PutUint32(frame[20:24], uint32(r.result))
	} else {
		frame[20] = byte(r.result)
	}
	return frame
}

func (r *DeliveryRsp) Decode(seq uint32, frame []byte) error {
	// check
	r.TotalLength = codec.HeadLen + uint32(len(frame))
	r.CommandId = CMPP_DELIVER_RESP
	r.SequenceId = seq
	r.msgId = binary.BigEndian.Uint64(frame[0:8])
	if V30.MajorMatchV(r.Version) {
		r.result = DlyResult(binary.BigEndian.Uint32(frame[8:12]))
	} else {
		r.result = DlyResult(frame[8])
	}
	return nil
}

func (r *DeliveryRsp) Log() (rt []log.Field) {
	rt = r.MessageHeader.Log()
	rt = append(rt,
		log.String("version", hex.EncodeToString([]byte{byte(r.Version)})),
		log.String("msgId", utils.Uint64HexString(r.msgId)),
		log.String("status", r.result.String()),
	)
	return
}

func (r *DeliveryRsp) SetResult(result DlyResult) {
	r.result = result
}

type DlyResult uint32

func (i DlyResult) String() string {
	return fmt.Sprintf("%d: %s", i, DeliveryResultMap[uint32(i)])
}

var DeliveryResultMap = map[uint32]string{
	0: "正确",
	1: "消息结构错",
	2: "命令字错",
	3: "消息序号重复",
	4: "消息长度错",
	5: "资费代码错",
	6: "超过最大信息长",
	7: "业务代码错",
	8: "流量控制错",
	9: "未知错误",
}
