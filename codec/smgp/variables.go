package smgp

import (
	"fmt"

	"github.com/hrygo/log"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type Version uint8

const (
	V30 Version = 0x30
	V20 Version = 0x20
	V13 Version = 0x13

	TP_pid           = uint16(0x0001)
	TP_udhi          = uint16(0x0002)
	PkTotal          = uint16(0x0009)
	PkNumber         = uint16(0x000A)
	LinkID           = uint16(0x0003)
	ChargeUserType   = uint16(0x0004)
	ChargeTermType   = uint16(0x0005)
	ChargeTermPseudo = uint16(0x0006)
	DestTermType     = uint16(0x0007)
	DestTermPseudo   = uint16(0x0008)
	SubmitMsgType    = uint16(0x000B)
	SPDealReslt      = uint16(0x000C)
	SrcTermType      = uint16(0x000D)
	SrcTermPseudo    = uint16(0x000E)
	NodesCount       = uint16(0x000F)
	MsgSrc           = uint16(0x0010)
	SrcType          = uint16(0x0011)
	MServiceID       = uint16(0x0012)
)

var (
	GbEncoder = simplifiedchinese.GB18030.NewEncoder()
	GbDecoder = simplifiedchinese.GB18030.NewDecoder()
)

func (t Version) String() string {
	switch {
	case t == V30:
		return "smgp30"
	case t == V20:
		return "smgp20"
	case t == V13:
		return "smgp13"
	default:
		return "unknown"
	}
}

// MajorMatch 主版本相匹配
func (t Version) MajorMatch(v uint8) bool {
	return uint8(t)&0xf0 == v&0xf0
}

// MajorMatchV 主版本相匹配
func (t Version) MajorMatchV(v Version) bool {
	return uint8(t)&0xf0 == uint8(v)&0xf0
}

// CommandId 命令定义
type CommandId uint32

const (
	SMGP_REQUEST_MIN, SMGP_RESPONSE_MIN CommandId = iota, 0x80000000 + iota
	SMGP_LOGIN, SMGP_LOGIN_RESP
	SMGP_SUBMIT, SMGP_SUBMIT_RESP
	SMGP_DELIVER, SMGP_DELIVER_RESP
	SMGP_ACTIVE_TEST, SMGP_ACTIVE_TEST_RESP
	_, _
	SMGP_EXIT, SMGP_EXIT_RESP
	SMGP_REQUEST_MAX, SMGP_RESPONSE_MAX
)

func (id CommandId) ToInt() uint32 {
	return uint32(id)
}

func (id CommandId) String() string {
	if id > SMGP_REQUEST_MIN && id < SMGP_REQUEST_MAX {
		return []string{
			"SMGP_LOGIN",
			"SMGP_SUBMIT",
			"SMGP_DELIVER",
			"SMGP_ACTIVE_TEST",
			"UNKNOWN",
			"SMGP_EXIT",
		}[id-1]
	} else if id > SMGP_RESPONSE_MIN && id < SMGP_RESPONSE_MAX {
		return []string{
			"SMGP_LOGIN_RESP",
			"SMGP_SUBMIT_RESP",
			"SMGP_DELIVER_RESP",
			"SMGP_ACTIVE_TEST_RESP",
			"UNKNOWN",
			"SMGP_EXIT_RESP",
		}[id-0x80000001]
	}
	return "UNKNOWN"
}

func (id CommandId) OpLog() log.Field {
	return log.String("op", id.String())
}

// Status 状态码
type Status uint32

func (s Status) String() string {
	return fmt.Sprintf("%d: %s", s, StatMap[uint32(s)])
}

var StatMap = map[uint32]string{
	0:  "成功",
	1:  "系统忙",
	2:  "超过最大连接数",
	10: "消息结构错",
	11: "命令字错",
	12: "序列号重复",
	20: "IP地址错",
	21: "认证错",
	22: "版本太高",
	30: "非法消息类型（MsgType）",
	31: "非法优先级（LruPriority）",
	32: "非法资费类型（FeeType）",
	33: "非法资费代码（FeeCode）",
	34: "非法短消息格式（MsgFormat）",
	35: "非法时间格式",
	36: "非法短消息长度（MsgLength）",
	37: "有效期已过",
	38: "非法查询类别（QueryType）",
	39: "路由错误",
	40: "非法包月费/封顶费（FixedFee）",
	41: "非法更新类型（UpdateType）",
	42: "非法路由编号（RouteId）",
	43: "非法服务代码（ServiceId）",
	44: "非法有效期（ValidTime）",
	45: "非法定时发送时间（AtTime）",
	46: "非法发送用户号码（SrcTermId）",
	47: "非法接收用户号码（DestTermId）",
	48: "非法计费用户号码（ChargeTermId）",
	49: "非法SP服务代码（SPCode）",
	56: "非法源网关代码（SrcGatewayID）",
	57: "非法查询号码（QueryTermID）",
	58: "没有匹配路由",
	59: "非法SP类型（SPType）",
	60: "非法上一条路由编号（LastRouteID）",
	61: "非法路由类型（RouteType）",
	62: "非法目标网关代码（DestGatewayID）",
	63: "非法目标网关IP（DestGatewayIP）",
	64: "非法目标网关端口（DestGatewayPort）",
	65: "非法路由号码段（TermRangeID）",
	66: "非法终端所属省代码（ProvinceCode）",
	67: "非法用户类型（UserType）",
	68: "本节点不支持路由更新",
	69: "非法SP企业代码（SPID）",
	70: "非法SP接入类型（SPAccessType）",
	71: "路由信息更新失败",
	72: "非法时间戳（Time）",
	73: "非法业务代码（MServiceID）",
	74: "SP禁止下发时段",
	75: "SP发送超过日流量",
	76: "SP帐号过有效期",
}
