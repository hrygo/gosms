package sgip

import (
	"encoding/binary"
	"encoding/hex"
	"strings"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/utils"
)

type Submit struct {
	MessageHeader
	SPNumber         string   //  SP的接入号码【 21 bytes 】
	ChargeNumber     string   //  付费号码，手机号码前加“86”国别标志；当且仅当群发且对用户收费时为空；如果为空，则该条短消息产生的费用由UserNumber代表的用户支付；如果为全零字符串“000000000000000000000”，表示该条短消息产生的费用由SP支付。【 21 bytes 】
	UserCount        byte     //  接收短消息的手机数量，取值范围1至100【 1  bytes 】
	UserNumber       []string //  接收该短消息的手机号，该字段重复UserCount指定的次数，手机号码前加“86”国别标志【 21 bytes 】
	CorpId           string   //  企业代码，取值范围0-99999 【 5  bytes 】
	ServiceType      string   //  业务代码，由SP定义 【 10 bytes 】
	FeeType          byte     //  计费类型【 1  bytes 】
	FeeValue         string   //  取值范围0-99999，该条短消息的收费值，单位为分，由SP定义,对于包月制收费的用户，该值为月租费的值 【 6  bytes 】
	GivenValue       string   //  取值范围0-99999，赠送用户的话费，单位为分，由SP定义，特指由SP向用户发送广告时的赠送话费【 6  bytes 】
	AgentFlag        byte     //  代收费标志，0：应收；1：实收 【 1  bytes 】
	MorelatetoMTFlag byte     //  引起MT消息的原因 【 1  bytes 】
	Priority         byte     //  优先级0-9从低到高，默认为0 【 1 bytes 】
	ExpireTime       string   //  短消息寿命的终止时间，如果为空，表示使用短消息中心的缺省值。时间内容为16个字符，格式为”yymmddhhmmsstnnp” ，其中“tnnp”取固定值“032+”，即默认系统为北京时间 【 16 bytes 】
	ScheduleTime     string   //  短消息定时发送的时间，如果为空，表示立刻发送该短消息。时间内容为16个字符，格式为“yymmddhhmmsstnnp” ，其中“tnnp”取固定值“032+”，即默认系统为北京时间 【 16  bytes 】
	ReportFlag       byte     //  状态报告标记【 1  bytes 】
	TpPid            byte     //  GSM协议类型。详细解释请参考GSM03.40中的9.2.3.9 【 1  bytes 】
	TpUdhi           byte     //  GSM协议类型。详细解释请参考GSM03.40中的9.2.3.9 【 1  bytes 】
	MessageCoding    byte     //  短消息的编码格式。 【 1  bytes 】
	MessageType      byte     //  信息类型: 0-短消息信息 其它:待定 【 1  bytes 】
	MessageLength    uint32   //  短消息的长度【 4  bytes 】
	MessageContent   []byte   //  编码后消息内容
	Reserve          string   //  保留，扩展用【 8  bytes 】

	// ReportFlag
	// 状态报告标记 0-该条消息只有最后出错时要返回状态报告 1-该条消息无论最后是否成功都要返回状态报告 2-该条消息不需要返回状态报告 3-该条消息仅携带包月计费信息，不下发给用户， 要返回状态报告
	// 其它-保留
	// 缺省设置为 0

	// MorelatetoMTFlag
	// 引起 MT 消息的原因
	// 0-MO 点播引起的第一条 MT 消息;
	// 1-MO 点播引起的非第一条 MT 消息;
	// 2-非 MO 点播引起的 MT 消息;
	// 3-系统反馈引起的 MT 消息。

	// MessageCoding
	// 短消息的编码格式。
	// 0:纯 ASCII 字符串
	// 3:写卡操作
	// 4:二进制编码
	// 8:UCS2 编码
	// 15: GBK 编码
	// 其它参见 GSM3.38 第 4 节:SMS Data Coding Scheme
}

const MtBaseLen = 143

func NewSubmit(ac *codec.AuthConf, phones []string, content string, options ...codec.OptionFunc) (messages []codec.RequestPdu) {
	mt := &Submit{}
	mt.PacketLength = MtBaseLen
	mt.CommandId = SGIP_SUBMIT
	mt.SequenceNumber = Sequencer.NextVal()
	mt.SetOptions(ac, codec.LoadMtOptions(options...))
	mt.UserCount = byte(len(phones))
	mt.UserNumber = phones
	mt.MessageCoding = utils.MsgFmt(content)
	mt.MorelatetoMTFlag = 2

	slices := utils.MsgSlices(mt.MessageCoding, content)
	if len(slices) == 1 {
		mt.MessageLength = uint32(len(slices[0]))
		mt.MessageContent = slices[0]
		mt.PacketLength = uint32(MtBaseLen + len(phones)*21 + len(slices[0]))
		return []codec.RequestPdu{mt}
	} else {
		mt.TpUdhi = 1
		for i, msgBytes := range slices {
			// 拷贝 mt
			tmp := *mt
			sub := &tmp
			if i != 0 {
				sub.SequenceNumber = Sequencer.NextVal()
			}
			sub.MessageLength = uint32(len(msgBytes))
			sub.MessageContent = msgBytes
			sub.PacketLength = uint32(MtBaseLen + len(phones)*21 + len(msgBytes))
			messages = append(messages, sub)
		}
		return messages
	}
}

func (s *Submit) Encode() []byte {
	frame := s.MessageHeader.Encode()
	index := 20
	copy(frame[index:], s.SPNumber)
	index += 21
	copy(frame[index:], s.ChargeNumber)
	index += 21
	frame[index] = s.UserCount
	index++
	for _, p := range s.UserNumber {
		copy(frame[index:], p)
		index += 21
	}
	copy(frame[index:], s.CorpId)
	index += 5
	copy(frame[index:], s.ServiceType)
	index += 10
	frame[index] = s.FeeType
	index++
	copy(frame[index:], s.FeeValue)
	index += 6
	copy(frame[index:], s.GivenValue)
	index += 6
	frame[index] = s.AgentFlag
	index++
	frame[index] = s.MorelatetoMTFlag
	index++
	frame[index] = s.Priority
	index++
	copy(frame[index:], s.ExpireTime)
	index += 16
	copy(frame[index:], s.ScheduleTime)
	index += 16
	frame[index] = s.ReportFlag
	index++
	frame[index] = s.TpPid
	index++
	frame[index] = s.TpUdhi
	index++
	frame[index] = s.MessageCoding
	index++
	frame[index] = s.MessageType
	index++
	binary.BigEndian.PutUint32(frame[index:], s.MessageLength)
	index += 4
	copy(frame[index:], s.MessageContent)
	index += len(s.MessageContent)
	copy(frame[index:], s.Reserve)
	return frame
}

func (s *Submit) Decode(cid uint32, frame []byte) error {
	s.PacketLength = codec.HeadLen + uint32(len(frame))
	s.CommandId = SGIP_SUBMIT
	s.SequenceNumber = make([]uint32, 3)
	s.SequenceNumber[0] = cid
	index := 0
	s.SequenceNumber[1] = binary.BigEndian.Uint32(frame[index:])
	index += 4
	s.SequenceNumber[2] = binary.BigEndian.Uint32(frame[index:])
	index += 4
	s.SPNumber = utils.TrimStr(frame[index : index+21])
	index += 21
	s.ChargeNumber = utils.TrimStr(frame[index : index+21])
	index += 21
	s.UserCount = frame[index]
	index++
	s.UserNumber = make([]string, s.UserCount)
	for i := 0; i < int(s.UserCount); i++ {
		s.UserNumber[i] = utils.TrimStr(frame[index : index+21])
		index += 21
	}
	s.CorpId = utils.TrimStr(frame[index : index+5])
	index += 5
	s.ServiceType = utils.TrimStr(frame[index : index+10])
	index += 10
	s.FeeType = frame[index]
	index++
	s.FeeValue = utils.TrimStr(frame[index : index+6])
	index += 6
	s.GivenValue = utils.TrimStr(frame[index : index+6])
	index += 6
	s.AgentFlag = frame[index]
	index++
	s.MorelatetoMTFlag = frame[index]
	index++
	s.Priority = frame[index]
	index++
	s.ExpireTime = utils.TrimStr(frame[index : index+16])
	index += 16
	s.ScheduleTime = utils.TrimStr(frame[index : index+16])
	index += 16
	s.ReportFlag = frame[index]
	index++
	s.TpPid = frame[index]
	index++
	s.TpUdhi = frame[index]
	index++
	s.MessageCoding = frame[index]
	index++
	s.MessageType = frame[index]
	index++
	s.MessageLength = binary.BigEndian.Uint32(frame[index:])
	index += 4
	content := frame[index : index+int(s.MessageLength)]
	s.MessageContent = content
	if content[0] == 0x05 && content[1] == 0x00 && content[2] == 0x03 {
		content = content[6:]
		s.MessageContent, _ = utils.Ucs2ToUtf8(content)
	}
	s.Reserve = ""
	return nil
}

func (s *Submit) SetOptions(ac *codec.AuthConf, ops *codec.MtOptions) {
	s.SPNumber = ac.SmsDisplayNo
	if ops.SpSubNo != "" {
		s.SPNumber += ops.SpSubNo
	}

	if len(ac.ClientId) > 5 {
		s.CorpId = ac.ClientId[5:]
	} else {
		s.CorpId = ac.ClientId
	}

	if ops.MsgLevel != uint8(0xf) {
		s.Priority = ops.MsgLevel
	} else {
		s.Priority = ac.DefaultMsgLevel
	}

	if ops.NeedReport != uint8(0xf) {
		s.ReportFlag = ops.NeedReport
	} else {
		s.ReportFlag = ac.NeedReport
	}

	s.ServiceType = ac.ServiceId
	if ops.ServiceId != "" {
		s.ServiceType = ops.ServiceId
	}

	if ops.AtTime != "" {
		s.ScheduleTime = ops.AtTime
	}

	if ops.ValidTime != "" {
		s.ExpireTime = ops.ValidTime
	} else {
		t := time.Now().Add(ac.MtValidDuration)
		s.ExpireTime = utils.FormatTime(t)
	}
}

func (s *Submit) Log() []log.Field {
	ls := s.MessageHeader.Log()
	var l = len(s.MessageContent)
	if l > 6 {
		l = 6
	}
	msg := hex.EncodeToString(s.MessageContent[:l]) + "..."
	return append(ls,
		log.String("spNumber", s.SPNumber),
		log.String("clientId", s.CorpId),
		log.Uint8("priority", s.Priority),
		log.Uint8("needReport", s.ReportFlag),
		log.String("serviceId", s.ServiceType),
		log.String("validTime", s.ExpireTime),
		log.String("atTime", s.ScheduleTime),
		log.Uint8("userCount", s.UserCount),
		log.String("userNumber", strings.Join(s.UserNumber, ",")),
		log.Uint8("msgType", s.MessageType),
		log.Uint8("msgFormat", s.MessageCoding),
		log.Uint32("msgLength", s.MessageLength),
		log.String("msgContent", msg),
		log.Uint8("tpPid", s.TpPid),
		log.Uint8("tpUdhi", s.TpUdhi),
		log.String("chargeNumber", s.ChargeNumber),
		log.Uint8("feeType", s.FeeType),
		log.String("feeValue", s.FeeValue),
		log.String("givenValue", s.GivenValue),
		log.Uint8("agentFlag", s.AgentFlag),
		log.Uint8("morelatetoMTFlag", s.MorelatetoMTFlag),
	)
}

type SubmitRsp struct {
	MessageHeader
	Status  Status
	Reserve string
}

func (s *Submit) ToResponse(code uint32) codec.Pdu {
	rsp := &SubmitRsp{}
	rsp.PacketLength = codec.HeadLen + 8 + 1 + 8
	rsp.CommandId = SGIP_SUBMIT_RESP
	rsp.SequenceNumber = s.SequenceNumber
	rsp.Status = Status(code)
	rsp.Reserve = ""
	return rsp
}

func (r *SubmitRsp) Decode(cid uint32, frame []byte) error {
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

func (r *SubmitRsp) Encode() []byte {
	frame := r.MessageHeader.Encode()
	frame[20] = byte(r.Status)
	return frame
}

func (r *SubmitRsp) Log() []log.Field {
	ls := r.MessageHeader.Log()
	return append(ls, log.String("status", r.Status.String()))
}
