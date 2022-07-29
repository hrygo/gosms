package cmpp

import (
	"github.com/hrygo/log"
)

type Version uint8

const (
	V30 Version = 0x30
	V21 Version = 0x21
	V20 Version = 0x20
)

func (t Version) String() string {
	switch {
	case t == V30:
		return "cmpp30"
	case t == V21:
		return "cmpp21"
	case t == V20:
		return "cmpp20"
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
	CMPP_REQUEST_MIN, CMPP_RESPONSE_MIN CommandId = iota, 0x80000000 + iota
	CMPP_CONNECT, CMPP_CONNECT_RESP
	CMPP_TERMINATE, CMPP_TERMINATE_RESP
	_, _
	CMPP_SUBMIT, CMPP_SUBMIT_RESP
	CMPP_DELIVER, CMPP_DELIVER_RESP
	CMPP_QUERY, CMPP_QUERY_RESP
	CMPP_CANCEL, CMPP_CANCEL_RESP
	CMPP_ACTIVE_TEST, CMPP_ACTIVE_TEST_RESP
	CMPP_FWD, CMPP_FWD_RESP
	CMPP_REQUEST_MAX, CMPP_RESPONSE_MAX
)

func (id CommandId) ToInt() uint32 {
	return uint32(id)
}

func (id CommandId) String() string {
	if id > CMPP_REQUEST_MIN && id < CMPP_REQUEST_MAX {
		return []string{
			"CMPP_CONNECT",
			"CMPP_TERMINATE",
			"CMPP_UNKNOWN",
			"CMPP_SUBMIT",
			"CMPP_DELIVER",
			"CMPP_QUERY",
			"CMPP_CANCEL",
			"CMPP_ACTIVE_TEST",
			"CMPP_FWD",
		}[id-1]
	} else if id > CMPP_RESPONSE_MIN && id < CMPP_RESPONSE_MAX {
		return []string{
			"CMPP_CONNECT_RESP",
			"CMPP_TERMINATE_RESP",
			"CMPP_UNKNOWN",
			"CMPP_SUBMIT_RESP",
			"CMPP_DELIVER_RESP",
			"CMPP_QUERY_RESP",
			"CMPP_CANCEL_RESP",
			"CMPP_ACTIVE_TEST_RESP",
			"CMPP_FWD_RESP",
		}[id-0x80000001]
	}
	return "UNKNOWN"
}

func (id CommandId) OpLog() log.Field {
	return log.String("op", id.String())
}
