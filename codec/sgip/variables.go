package sgip

import (
	"fmt"

	"github.com/hrygo/log"
)

type Version uint8

const V12 Version = 0x12

func (t Version) String() string {
	switch {
	case t == V12:
		return "sgip12"
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
	SGIP_REQUEST_MIN, SGIP_RESPONSE_MIN CommandId = iota, 0x80000000 + iota
	SGIP_BIND, SGIP_BIND_RESP
	SGIP_UNBIND, SGIP_UNBIND_RESP
	SGIP_SUBMIT, SGIP_SUBMIT_RESP
	SGIP_DELIVER, SGIP_DELIVER_RESP
	SGIP_REPORT, SGIP_REPORT_RESP
	SGIP_REQUEST_MAX, SGIP_RESPONSE_MAX
)

func (id CommandId) ToInt() uint32 {
	return uint32(id)
}

func (id CommandId) String() string {
	if id > SGIP_REQUEST_MIN && id < SGIP_REQUEST_MAX {
		return []string{
			"SGIP_BIND",
			"SGIP_UNBIND",
			"SGIP_SUBMIT",
			"SGIP_DELIVER",
			"SGIP_REPORT",
		}[id-1]
	} else if id > SGIP_RESPONSE_MIN && id < SGIP_RESPONSE_MAX {
		return []string{
			"SGIP_BIND_RESP",
			"SGIP_UNBIND_RESP",
			"SGIP_SUBMIT_RESP",
			"SGIP_DELIVER_RESP",
			"SGIP_REPORT_RESP",
		}[id-0x80000001]
	}
	return "UNKNOWN"
}

func (id CommandId) OpLog() log.Field {
	return log.String("op", id.String())
}

type Status byte

func (r Status) String() string {
	return fmt.Sprintf("%d: %s", r, ResultMap[r])
}

var ResultMap = map[Status]string{
	0:  "成功",
	1:  "非法登录, 如登录名、口令出错、登录名与口令不符等。",
	2:  "重复登录, 如在同一TCP/IP连接中连续两次以上请求登录。",
	3:  "连接过多, 指单个节点要求同时建立的连接数过多。",
	4:  "登录类型错, 指bind命令中的logintype字段出错。",
	5:  "参数格式错, 指命令中参数值与参数类型不符或与协议规定的范围不符。",
	6:  "非法手机号码, 协议中所有手机号码字段出现非86130号码或手机号码前未加“86”时都应报错。",
	7:  "消息ID错",
	8:  "信息长度错",
	9:  "非法序列号, 包括序列号重复、序列号格式错误等",
	10: "非法操作GNS",
	11: "节点忙, 指本节点存储队列满或其他原因, 暂时不能提供服务的情况",
	21: "目的地址不可达, 指路由表存在路由且消息路由正确但被路由的节点暂时不能提供服务的情况",
	22: "路由错, 指路由表存在路由但消息路由出错的情况, 如转错SMG等",
	23: "路由不存在, 指消息路由的节点在路由表中不存在",
	24: "计费号码无效, 鉴权不成功时反馈的错误信息",
	25: "用户不能通信（如不在服务区、未开机等情况）",
	26: "手机内存不足",
	27: "手机不支持短消息",
	28: "手机接收短消息出现错误",
	29: "不知道的用户",
	30: "不提供此功能",
	31: "非法设备",
	32: "系统失败",
	33: "短信中心队列满",
}
