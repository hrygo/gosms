package cmpp

import (
	"encoding/binary"
	"encoding/hex"
	"strings"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/client"
	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/utils"
)

// Submit
// 3.0版 feeTerminalId、destTerminalId 均为32字节，无Reserve字段，有LinkId字段
// 2.0版 feeTerminalId、destTerminalId 均为21字节，无LinkId字段，有Reserve字段

type Submit struct {
	MessageHeader           // 消息头，【12字节】
	msgId            uint64 // 信息标识，由 SP 接入的短信网关本身产 生，本处填空(0)。【8字节】
	pkTotal          uint8  // 相同Msg_Id的信息总条数 【1字节】
	pkNumber         uint8  // 相同Msg_Id的信息序号，从1开始 【1字节】
	registeredDel    uint8  // 是否要求返回状态确认报告： 0：不需要，1：需要。【1字节】
	msgLevel         uint8  // 信息级别，1-9 【1字节】
	serviceId        string // 业务标识，是数字、字母和符号的组合。【10字节】
	feeUsertype      uint8  // 计费用户类型字段 【1字节】
	feeTerminalId    string //  被计费用户的号码（如本字节填空，则表示本字段无效，对谁计费参见Fee_UserType字段，本字段与Fee_UserType字段互斥）【32字节】
	feeTerminalType  uint8  // 被计费用户的号码类型，0：真实号码；1：伪码 【1字节】
	tpPid            uint8  // GSM协议类型。详细是解释请参考GSM03.40中的9.2.3.9 【1字节】
	tpUdhi           uint8  // GSM协议类型。详细是解释请参考GSM03.40中的9.2.3.9 【1字节】
	msgFmt           uint8  // 信息格式 【1字节】
	msgSrc           string // 信息内容来源(SP_Id) 【6字节】
	feeType          string //  资费类别【2字节】
	feeCode          string // 资费代码（以分为单位） 【6字节】
	validTime        string // 存活有效期，格式遵循SMPP3.3协议 【17字节】
	atTime           string // 定时发送时间，格式遵循SMPP3.3协议 【17字节】
	srcId            string //  源号码 SP的服务代码或前缀为服务代码的长号码, 网关将该号码完整的填到SMPP协议Submit_SM消息相应的source_addr字段，该号码最终在用户手机上显示为短消息的主叫号码【21字节】
	destUsrTl        uint8  // 接收信息的用户数量(小于100个用户) 【1字节】
	destTerminalId   string //  接收短信的MSISDN号码【32*DestUsrTl字节】
	termIds          []byte // DestTerminalId编码后的格式
	destTerminalType uint8  //  接收短信的用户的号码类型，0：真实号码；1：伪码【1字节】
	msgLength        uint8  // 信息长度(Msg_Fmt值为0时：<160个字节；其它<=140个字节) 【1字节】
	msgContent       string // 信息内容 【MsgLength字节】
	msgBytes         []byte // 消息内容按照Msg_Fmt编码后的数据
	linkID           string // 点播业务使用的LinkID，非点播类业务的MT流程不使用该字段 【20字节】

	// 协议版本,不是报文内容，但在调用encode方法前需要设置此值
	Version Version
}

func NewSubmit(cli *client.Client, phones []string, content string, seq uint32, opts ...Option) (messages []codec.RequestPdu) {
	options := loadOptions(opts...)
	baseLen := 138
	if V30.MajorMatch(cli.Version) {
		baseLen = 163
	}
	header := MessageHeader{TotalLength: uint32(baseLen), CommandId: CMPP_SUBMIT, SequenceId: seq}
	mt := &Submit{MessageHeader: header, Version: Version(cli.Version)}

	setOptions(cli, mt, options)
	mt.msgFmt = MsgFmt(content)

	mt.destUsrTl = uint8(len(phones))
	mt.destTerminalId = strings.Join(phones, ",")
	idLen := 21
	if V30.MajorMatch(cli.Version) {
		idLen = 32
	}
	termIds := make([]byte, idLen*int(mt.destUsrTl))
	for i, p := range phones {
		copy(termIds[i*idLen:(i+1)*idLen], p)
	}
	mt.termIds = termIds

	mt.msgSrc = cli.SmsDisplayNo

	mt.msgContent = content
	slices := MsgSlices(mt.msgFmt, content)

	if len(slices) == 1 {
		mt.pkTotal = 1
		mt.pkNumber = 1
		mt.msgLength = uint8(len(slices[0]))
		mt.msgBytes = slices[0]
		mt.TotalLength = uint32(baseLen + len(termIds) + len(slices[0]))
		return []codec.RequestPdu{mt}
	} else {
		mt.tpUdhi = 1
		mt.pkTotal = uint8(len(slices))

		for i, msgBytes := range slices {
			// 拷贝 mt
			tmp := *mt
			sub := &tmp
			if i != 0 {
				sub.SequenceId = uint32(codec.B32Seq.NextVal())
			}
			sub.pkNumber = uint8(i + 1)
			sub.msgLength = uint8(len(msgBytes))
			sub.msgBytes = msgBytes
			sub.TotalLength = uint32(baseLen + len(termIds) + len(msgBytes))
			messages = append(messages, sub)
		}

		return messages
	}
}

func (s *Submit) Encode() []byte {
	frame := s.MessageHeader.Encode()
	frame[20] = s.pkTotal
	frame[21] = s.pkNumber
	frame[22] = s.registeredDel
	frame[23] = s.msgLevel
	copy(frame[24:34], s.serviceId)
	frame[34] = s.feeUsertype
	index := 35
	if V30.MajorMatchV(s.Version) {
		copy(frame[index:index+32], s.feeTerminalId)
		index += 32
		frame[index] = s.feeTerminalType
		index++
	} else {
		copy(frame[index:index+21], s.feeTerminalId)
		index += 21
	}
	frame[index] = s.tpPid
	index++
	frame[index] = s.tpUdhi
	index++
	frame[index] = s.msgFmt
	index++
	copy(frame[index:index+6], s.msgSrc)
	index += 6
	copy(frame[index:index+2], s.feeType)
	index += 2
	copy(frame[index:index+6], s.feeCode)
	index += 6
	copy(frame[index:index+17], s.validTime)
	index += 17
	copy(frame[index:index+17], s.atTime)
	index += 17
	copy(frame[index:index+21], s.srcId)
	index += 21
	frame[index] = s.destUsrTl
	index++
	copy(frame[index:index+len(s.termIds)], s.termIds)
	index += len(s.termIds)
	if V30.MajorMatchV(s.Version) {
		frame[index] = s.destTerminalType
		index++
	}
	frame[index] = s.msgLength
	index++
	copy(frame[index:index+len(s.msgBytes)], s.msgBytes)
	index += len(s.msgBytes)
	if V30.MajorMatchV(s.Version) {
		copy(frame[index:index+20], s.linkID)
	}
	return frame
}

func (s *Submit) Decode(seq uint32, frame []byte) error {
	s.TotalLength = HeadLen + uint32(len(frame))
	s.CommandId = CMPP_SUBMIT
	s.SequenceId = seq
	// msgId uint64
	index := 8
	s.pkTotal = frame[index]
	index++
	s.pkNumber = frame[index]
	index++
	s.registeredDel = frame[index]
	index++
	s.msgLevel = frame[index]
	index++
	s.serviceId = utils.TrimStr(frame[index : index+10])
	index += 10
	s.feeUsertype = frame[index]
	index++
	if V30.MajorMatchV(s.Version) {
		s.feeTerminalId = utils.TrimStr(frame[index : index+32])
		index += 32
		s.feeTerminalType = frame[index]
		index++
	} else {
		s.feeTerminalId = utils.TrimStr(frame[index : index+21])
		index += 21
	}
	s.tpPid = frame[index]
	index++
	s.tpUdhi = frame[index]
	index++
	s.msgFmt = frame[index]
	index++
	s.msgSrc = utils.TrimStr(frame[index : index+6])
	index += 6
	s.feeType = utils.TrimStr(frame[index : index+2])
	index += 2
	s.feeCode = utils.TrimStr(frame[index : index+6])
	index += 6
	s.validTime = utils.TrimStr(frame[index : index+17])
	index += 17
	s.atTime = utils.TrimStr(frame[index : index+17])
	index += 17
	s.srcId = utils.TrimStr(frame[index : index+21])
	index += 21
	s.destUsrTl = frame[index]
	index++
	l := int(s.destUsrTl * 21)
	if V30.MajorMatchV(s.Version) {
		l = int(s.destUsrTl) << 5
	}
	s.termIds = frame[index : index+l]
	index += l
	if V30.MajorMatchV(s.Version) {
		s.destTerminalType = frame[index]
		index++
	}
	s.msgLength = frame[index]
	index++
	content := frame[index : index+int(s.msgLength)]
	s.msgBytes = content
	if content[0] == 0x05 && content[1] == 0x00 && content[2] == 0x03 {
		content = content[6:]
	}
	if s.msgFmt == 8 {
		s.msgContent, _ = utils.Ucs2ToUtf8(content)
	} else {
		s.msgContent = string(content)
	}
	index += int(s.msgLength)
	if V30.MajorMatchV(s.Version) {
		s.linkID = utils.TrimStr(frame[index : index+20])
	}
	return nil
}

type SubmitResp struct {
	MessageHeader
	msgId  uint64
	result MtResult

	// 协议版本,不是报文内容，但在调用encode方法前需要设置此值
	Version Version
}

func (s *Submit) ToResponse(result uint32) codec.Pdu {
	resp := &SubmitResp{}
	resp.TotalLength = HeadLen + 9
	resp.CommandId = CMPP_SUBMIT_RESP
	resp.SequenceId = s.SequenceId
	if V30.MajorMatchV(s.Version) {
		resp.TotalLength = HeadLen + 12
	}
	if result == 0 {
		resp.msgId = uint64(codec.B64Seq.NextVal())
	}
	resp.result = MtResult(result)

	resp.Version = s.Version
	return resp
}

func (s *Submit) ToDeliveryReport(msgId uint64) *Delivery {
	d := Delivery{}
	d.TotalLength = 145
	if V30.MajorMatchV(s.Version) {
		d.TotalLength = 169
	}
	d.CommandId = CMPP_DELIVER
	d.SequenceId = uint32(codec.B32Seq.NextVal())

	d.registeredDelivery = 1
	d.msgLength = 60
	d.destId = s.srcId
	d.serviceId = s.serviceId
	d.srcTerminalId = s.destTerminalId
	d.srcTerminalType = s.destTerminalType

	subTime := time.Now().Format("0601021504")
	doneTime := time.Now().Add(10 * time.Second).Format("0601021504")
	report := NewReport(msgId, s.destTerminalId, subTime, doneTime)
	d.report = report

	return &d
}

func (r *SubmitResp) Encode() []byte {
	frame := r.MessageHeader.Encode()
	binary.BigEndian.PutUint64(frame[12:20], r.msgId)
	if V30.MajorMatchV(r.Version) {
		binary.BigEndian.PutUint32(frame[20:24], uint32(r.result))
	} else {
		frame[20] = byte(r.result)
	}
	return frame
}
func (r *SubmitResp) Decode(seq uint32, frame []byte) error {
	r.TotalLength = HeadLen + uint32(len(frame))
	r.CommandId = CMPP_SUBMIT_RESP
	r.SequenceId = seq
	r.msgId = binary.BigEndian.Uint64(frame[0:8])
	if V30.MajorMatchV(r.Version) {
		r.result = MtResult(binary.BigEndian.Uint32(frame[8:12]))
	} else {
		r.result = MtResult(frame[8])
	}
	return nil
}

func (r *SubmitResp) MsgId() uint64 {
	return r.msgId
}

func (r *SubmitResp) Result() uint32 {
	return uint32(r.result)
}

func (r *SubmitResp) Log() (rt []log.Field) {
	rt = r.MessageHeader.Log()
	rt = append(rt,
		log.Uint8("version", uint8(r.Version)),
		log.Uint64("msgId", r.msgId),
		log.Uint8("result", uint8(r.result)),
	)
	return
}
func MsgSlices(fmt uint8, content string) (slices [][]byte) {
	var msgBytes []byte
	// 含中文
	if fmt == 8 {
		msgBytes, _ = utils.Utf8ToUcs2(content)
		slices = utils.ToTPUDHISlices(msgBytes, 140)
	} else {
		// 纯英文
		msgBytes = []byte(content)
		slices = utils.ToTPUDHISlices(msgBytes, 160)
	}
	return
}

// MsgFmt 通过消息内容判断，设置编码格式。
// 如果是纯拉丁字符采用0：ASCII串
// 如果含多字节字符，这采用8：UCS-2编码
func MsgFmt(content string) uint8 {
	if len(content) < 2 {
		return 0
	}
	all7bits := len(content) == len([]rune(content))
	if all7bits {
		return 0
	} else {
		return 8
	}
}

func (s *Submit) Log() []log.Field {
	var pl = 21
	if V30.MajorMatchV(s.Version) {
		pl = 32
	}
	ls := s.MessageHeader.Log()
	ls = append(ls,
		log.Uint8("version", uint8(s.Version)),
		log.Uint8("pkTotal", s.pkTotal),
		log.Uint8("pkNumber", s.pkNumber),
		log.Uint8("needReport", s.registeredDel),
		log.String("clientId", s.msgSrc),
		log.String("serviceId", s.serviceId),
		log.Uint8("tpPid", s.tpPid),
		log.Uint8("tpUdhi", s.tpUdhi),
		log.String("validTime", s.validTime),
		log.String("atTime", s.atTime),
		log.String("srcId", s.srcId),
		log.Uint8("destUsrTl", s.destUsrTl),
		log.String("destTerminalId", strings.Join(bytes2StringSlice(s.termIds, pl), ",")),
		log.Uint8("destTerminalType", s.destTerminalType),
		log.Uint8("msgLevel", s.msgLevel),
		log.Uint8("msgFmt", s.msgFmt),
		log.Uint8("msgLength", s.msgLength),
		log.String("linkID", s.linkID),
		log.Uint8("feeUsertype", s.feeUsertype),
		log.String("feeTerminalId", s.feeTerminalId),
		log.Uint8("feeTerminalType", s.feeTerminalType),
		log.String("feeType", s.feeType),
		log.String("feeCode", s.feeCode),
	)
	var csl log.Field
	var l = len(s.msgBytes)
	if l > 6 {
		l = 6
	}
	csl = log.String("msgContent", hex.EncodeToString(s.msgBytes[:l])+"...")
	return append(ls, csl)
}

// 设置可选项
func setOptions(cli *client.Client, sub *Submit, opts *MtOptions) {
	if opts.FeeUsertype != uint8(0xf) {
		sub.feeUsertype = opts.FeeUsertype
	} else {
		sub.feeUsertype = cli.FeeUserType
	}

	if opts.MsgLevel != uint8(0xf) {
		sub.msgLevel = opts.MsgLevel
	} else {
		sub.msgLevel = cli.DefaultMsgLevel
	}

	if opts.RegisteredDel != uint8(0xf) {
		sub.registeredDel = opts.RegisteredDel
	} else {
		sub.registeredDel = cli.NeedReport
	}

	if opts.FeeTerminalType != uint8(0xf) {
		sub.feeTerminalType = opts.FeeTerminalType
	} else {
		sub.feeTerminalType = cli.FeeTerminalType
	}

	if opts.FeeType != "" {
		sub.feeType = opts.FeeType
	} else {
		sub.feeType = cli.FeeType
	}

	if opts.AtTime != "" {
		sub.atTime = opts.AtTime
	}

	if opts.ValidTime != "" {
		sub.validTime = opts.ValidTime
	} else {
		t := time.Now().Add(cli.MtValidDuration)
		s := t.Format("060102150405")
		sub.validTime = s + "032+"
	}

	if opts.FeeCode != "" {
		sub.feeCode = opts.FeeCode
	} else {
		sub.feeCode = cli.FeeCode
	}

	if opts.FeeTerminalId != "" {
		sub.feeTerminalId = opts.FeeTerminalId
	} else {
		sub.feeTerminalId = cli.FeeTerminalId
	}

	if opts.SrcId != "" {
		sub.srcId = opts.SrcId
	} else {
		sub.srcId = cli.SmsDisplayNo
	}

	if opts.ServiceId != "" {
		sub.serviceId = opts.ServiceId
	} else {
		sub.serviceId = cli.ServiceId
	}

	if opts.LinkID != "" {
		sub.linkID = opts.LinkID
	} else {
		sub.linkID = cli.LinkId
	}
}

type MtResult uint32

const (
	MtStatusOK MtResult = iota
	MtRsp1
	MtRsp2
	MtRsp3
	MtRsp4
	MtRsp5
	MtRsp6
	MtRsp7
	MtRsp8
	MtRsp9
	MtRsp10
	MtRsp11
	MtRsp12
	MtRsp13
)

func (i MtResult) String() string {
	return []string{
		"正确",
		"消息结构错",
		"命令字错",
		"消息序号重复",
		"消息长度错",
		"资费代码错",
		"超过最大信息长",
		"业务代码错",
		"流量控制错",
		"本网关不负责服务此计费号码",
		"Src_Id 错误",
		"Msg_src 错误",
		"Fee_terminal_Id 错误",
		"Dest_terminal_Id 错误",
	}[i]
}

func bytes2StringSlice(in []byte, pl int) (ret []string) {
	if len(in) <= pl {
		return []string{utils.TrimStr(in)}
	} else {
		part := len(in) / pl
		ret = make([]string, part)
		for i := 0; i < part; i++ {
			ret[i] = utils.TrimStr(in[i*pl : (i+1)*pl])
		}
	}
	return
}
