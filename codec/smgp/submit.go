package smgp

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/utils"
)

type Submit struct {
	MessageHeader
	msgType         byte           // 【1字节】短消息类型
	needReport      byte           // 【1字节】SP是否要求返回状态报告
	priority        byte           // 【1字节】短消息发送优先级
	serviceID       string         // 【10字节】业务代码
	feeType         string         // 【2字节】收费类型
	feeCode         string         // 【6字节】资费代码
	fixedFee        string         // 【6字节】包月费/封顶费
	msgFormat       byte           // 【1字节】短消息格式
	validTime       string         // 【17字节】短消息有效时间
	atTime          string         // 【17字节】短消息定时发送时间
	srcTermID       string         // 【21字节】短信息发送方号码
	chargeTermID    string         // 【21字节】计费用户号码
	destTermIDCount byte           // 【1字节】短消息接收号码总数
	destTermID      []string       // 【21*DestTermCount字节】短消息接收号码
	msgLength       byte           // 【1字节】短消息长度
	msgContent      string         // 【MsgLength字节】短消息内容
	msgBytes        []byte         // 消息内容按照Msg_Fmt编码后的数据
	reserve         string         // 【8字节】保留
	tlvList         *utils.TlvList // 【TLV】可选项参数

	// 协议版本,不是报文内容，但在调用encode方法前需要设置此值
	Version Version
}

type SubmitRsp struct {
	MessageHeader
	msgId  []byte // 【10字节】短消息流水号
	status Status

	// 协议版本,不是报文内容，但在调用encode方法前需要设置此值
	Version Version
}

const MtBaseLen = 126

func NewSubmit(ac *codec.AuthConf, phones []string, content string, seq uint32, options ...codec.OptionFunc) (messages []codec.RequestPdu) {
	mt := &Submit{Version: Version(ac.Version)}
	mt.PacketLength = MtBaseLen
	mt.RequestId = SMGP_SUBMIT
	mt.SequenceId = seq
	mt.SetOptions(ac, codec.LoadMtOptions(options...))
	mt.msgType = 6
	// 从配置文件设置属性
	mt.feeType = ac.FeeType
	mt.feeCode = ac.FeeCode
	mt.chargeTermID = ac.FeeTerminalId
	mt.fixedFee = ac.FixedFee
	// 初步设置入参
	mt.destTermID = phones
	mt.destTermIDCount = byte(len(phones))

	mt.msgFormat = 15
	data, err := GbEncoder.Bytes([]byte(content))
	if err != nil {
		return nil
	}
	slices := utils.ToTPUDHISlices(data, 140)
	if len(slices) == 1 {
		mt.msgBytes = slices[0]
		mt.msgLength = byte(len(mt.msgBytes))
		mt.PacketLength = uint32(MtBaseLen + len(mt.destTermID)*21 + int(mt.msgLength))
		return []codec.RequestPdu{mt}
	} else {
		for i, dt := range slices {
			// 拷贝 mt
			tmp := *mt
			sub := &tmp
			if i != 0 {
				sub.SequenceId = uint32(codec.B32Seq.NextVal())
			}
			sub.msgLength = byte(len(dt))
			sub.msgBytes = dt
			l := 0
			sub.tlvList = utils.NewTlvList()
			sub.tlvList.Add(TP_pid, []byte{0x00})
			l += 5
			sub.tlvList.Add(TP_udhi, []byte{0x01})
			l += 5
			sub.tlvList.Add(PkTotal, []byte{byte(len(slices))})
			l += 5
			sub.tlvList.Add(PkNumber, []byte{byte(i)})
			l += 5
			sub.PacketLength = uint32(MtBaseLen + len(sub.destTermID)*21 + int(sub.msgLength) + l)
			messages = append(messages, sub)
		}
		return messages
	}
}

func (s *Submit) Encode() []byte {
	if len(s.destTermID) != int(s.destTermIDCount) {
		return nil
	}
	frame := s.MessageHeader.Encode()
	index := 12
	index = utils.CopyByte(frame, s.msgType, index)
	index = utils.CopyByte(frame, s.needReport, index)
	index = utils.CopyByte(frame, s.priority, index)
	index = utils.CopyStr(frame, s.serviceID, index, 10)
	index = utils.CopyStr(frame, s.feeType, index, 2)
	index = utils.CopyStr(frame, s.feeCode, index, 6)
	index = utils.CopyStr(frame, s.fixedFee, index, 6)
	index = utils.CopyByte(frame, s.msgFormat, index)
	index = utils.CopyStr(frame, s.validTime, index, 17)
	index = utils.CopyStr(frame, s.atTime, index, 17)
	index = utils.CopyStr(frame, s.srcTermID, index, 21)
	index = utils.CopyStr(frame, s.chargeTermID, index, 21)
	index = utils.CopyByte(frame, s.destTermIDCount, index)
	for _, tid := range s.destTermID {
		index = utils.CopyStr(frame, tid, index, 21)
	}

	index = utils.CopyByte(frame, s.msgLength, index)
	copy(frame[index:index+int(s.msgLength)], s.msgBytes)
	index += +int(s.msgLength)
	index = utils.CopyStr(frame, s.reserve, index, 8)
	if s.tlvList != nil {
		buff := new(bytes.Buffer)
		err := s.tlvList.Write(buff)
		if err != nil {
			log.Errorf("%v", err)
			return nil
		}
		copy(frame[index:], buff.Bytes())
	}
	return frame
}

func (s *Submit) Decode(seq uint32, frame []byte) error {
	s.PacketLength = codec.HeadLen + uint32(len(frame))
	s.RequestId = SMGP_SUBMIT
	s.SequenceId = seq

	var index int
	s.msgType = frame[index]
	index++
	s.needReport = frame[index]
	index++
	s.priority = frame[index]
	index++
	s.serviceID = utils.TrimStr(frame[index : index+10])
	index += 10
	s.feeType = utils.TrimStr(frame[index : index+2])
	index += 2
	s.feeCode = utils.TrimStr(frame[index : index+6])
	index += 6
	s.fixedFee = utils.TrimStr(frame[index : index+6])
	index += 6
	s.msgFormat = frame[index]
	index++
	s.validTime = utils.TrimStr(frame[index : index+17])
	index += 17
	s.atTime = utils.TrimStr(frame[index : index+17])
	index += 17
	s.srcTermID = utils.TrimStr(frame[index : index+21])
	index += 21
	s.chargeTermID = utils.TrimStr(frame[index : index+21])
	index += 21
	s.destTermIDCount = frame[index]
	index++
	for i := byte(0); i < s.destTermIDCount; i++ {
		s.destTermID = append(s.destTermID, utils.TrimStr(frame[index:index+21]))
		index += 21
	}
	s.msgLength = frame[index]
	index++
	content := frame[index : index+int(s.msgLength)]
	s.msgBytes = content
	if content[0] == 0x05 && content[1] == 0x00 && content[2] == 0x03 {
		content = content[6:]
	}
	index += int(s.msgLength)
	tmp, _ := GbDecoder.Bytes(content)
	s.msgContent = string(tmp)
	s.reserve = utils.TrimStr(frame[index : index+8])
	index += 8
	// 一个tlv至少5字节
	if uint32(index+5) < s.PacketLength {
		buf := bytes.NewBuffer(frame[index:])
		s.tlvList, _ = utils.Read(buf)
	}
	return nil
}

func (s *Submit) ToResponse(code uint32) codec.Pdu {
	resp := &SubmitRsp{Version: s.Version}
	resp.PacketLength = 26
	resp.RequestId = SMGP_SUBMIT_RESP
	resp.SequenceId = s.SequenceId
	resp.status = Status(code)
	resp.msgId = codec.BcdSeq.NextVal()
	return resp
}

func (s *Submit) String() string {
	bts := s.msgBytes
	if s.msgLength > 6 {
		bts = s.msgBytes[:6]
	}
	return fmt.Sprintf("{ header: %s, msgType: %v, NeedReport: %v, LruPriority: %v, ServiceID: %v, "+
		"feeType: %v, feeCode: %v, fixedFee: %v, msgFormat: %v, validTime: %v, AtTime: %v, SrcTermID: %v, "+
		"chargeTermID: %v, destTermIDCount: %v, destTermID: %v, msgLength: %v, msgContent: %#x..., "+
		"reserve: %v, tlvList: %s }",
		&s.MessageHeader, s.msgType, s.needReport, s.priority, s.serviceID,
		s.feeType, s.feeCode, s.fixedFee, s.msgFormat, s.validTime, s.atTime, s.srcTermID,
		s.chargeTermID, s.destTermIDCount, s.destTermID, s.msgLength, bts,
		s.reserve, s.tlvList)
}

func (r *SubmitRsp) Encode() []byte {
	frame := r.MessageHeader.Encode()
	index := 12
	copy(frame[index:index+10], r.msgId)
	index += 10
	binary.BigEndian.PutUint32(frame[index:index+4], uint32(r.status))
	return frame
}

func (r *SubmitRsp) Decode(seq uint32, frame []byte) error {
	r.PacketLength = codec.HeadLen + uint32(len(frame))
	r.RequestId = SMGP_SUBMIT_RESP
	r.SequenceId = seq
	r.msgId = make([]byte, 10)
	copy(r.msgId, frame[0:10])
	r.status = Status(binary.BigEndian.Uint32(frame[10:14]))
	return nil
}

func (r *SubmitRsp) String() string {
	return fmt.Sprintf("{ header: %s, msgId: %x, status: \"%s\" }", &r.MessageHeader, r.msgId, r.status.String())
}

func (s *Submit) Log() []log.Field {
	ls := s.MessageHeader.Log()
	var l = len(s.msgBytes)
	if l > 6 {
		l = 6
	}
	msg := hex.EncodeToString(s.msgBytes[:l]) + "..."
	return append(ls,
		log.String("spNumber", s.srcTermID),
		log.Uint8("priority", s.priority),
		log.String("serviceId", s.serviceID),
		log.Uint8("needReport", s.needReport),
		log.String("validTime", s.validTime),
		log.String("atTime", s.atTime),
		log.Uint8("userCount", s.destTermIDCount),
		log.String("userNumber", strings.Join(s.destTermID, ",")),
		log.Uint8("msgType", s.msgType),
		log.Uint8("msgFormat", s.msgFormat),
		log.Uint8("msgLength", s.msgLength),
		log.String("msgContent", msg),
		log.String("feeType", s.feeType),
		log.String("feeCode", s.feeCode),
		log.String("fixedFee", s.fixedFee),
		log.String("chargeTermID", s.chargeTermID),
		log.String("version", hex.EncodeToString([]byte{byte(s.Version)})),
		log.String("tlv", s.tlvList.String()),
	)
}

func (r *SubmitRsp) Log() (rt []log.Field) {
	rt = r.MessageHeader.Log()
	rt = append(rt,
		log.String("version", hex.EncodeToString([]byte{byte(r.Version)})),
		log.String("msgId", hex.EncodeToString(r.msgId)),
		log.String("status", r.status.String()),
	)
	return
}

func (s *Submit) MsgType() byte {
	return s.msgType
}

func (s *Submit) NeedReport() byte {
	return s.needReport
}

func (s *Submit) Priority() byte {
	return s.priority
}

func (s *Submit) ServiceID() string {
	return s.serviceID
}

func (s *Submit) FeeType() string {
	return s.feeType
}

func (s *Submit) FeeCode() string {
	return s.feeCode
}

func (s *Submit) FixedFee() string {
	return s.fixedFee
}

func (s *Submit) MsgFormat() byte {
	return s.msgFormat
}

func (s *Submit) ValidTime() string {
	return s.validTime
}

func (s *Submit) AtTime() string {
	return s.atTime
}

func (s *Submit) SrcTermID() string {
	return s.srcTermID
}

func (s *Submit) ChargeTermID() string {
	return s.chargeTermID
}

func (s *Submit) DestTermIDCount() byte {
	return s.destTermIDCount
}

func (s *Submit) DestTermID() []string {
	return s.destTermID
}

func (s *Submit) MsgLength() byte {
	return s.msgLength
}

func (s *Submit) MsgContent() string {
	return s.msgContent
}

func (s *Submit) MsgBytes() []byte {
	return s.msgBytes
}

func (s *Submit) Reserve() string {
	return s.reserve
}

func (s *Submit) TlvList() *utils.TlvList {
	return s.tlvList
}

func (r *SubmitRsp) MsgId() []byte {
	return r.msgId
}

func (r *SubmitRsp) Status() Status {
	return r.status
}

func (s *Submit) SetOptions(ac *codec.AuthConf, options *codec.MtOptions) {
	s.needReport = ac.NeedReport
	// 有点小bug，不能通过传参的方式设置未变量的"零值"
	if options.NeedReport != 0 {
		s.needReport = options.NeedReport
	}

	s.priority = ac.DefaultMsgLevel
	// 有点小bug，不能通过传参的方式设置未变量的"零值"
	if options.MsgLevel != 0 {
		s.priority = options.MsgLevel
	}

	s.serviceID = ac.ServiceId
	if options.ServiceId != "" {
		s.serviceID = options.ServiceId
	}

	if options.AtTime != "" {
		s.atTime = options.AtTime
	}

	vt := time.Now()
	if options.ValidTime != "" {
		s.validTime = options.ValidTime
	} else {
		vt.Add(ac.MtValidDuration)
	}
	s.validTime = utils.FormatTime(vt)

	s.srcTermID = ac.SmsDisplayNo
	if options.SrcId != "" {
		s.srcTermID += options.SrcId
	}
}
