package cmpp

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

const (
	HeadLen     uint32 = 12
	PacketMin   uint32 = HeadLen
	PacketMaxV2 uint32 = 2477
	PacketMaxV3 uint32 = 3335
	PacketMax   uint32 = PacketMaxV3
)

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
	CMPP_MT_ROUTE, CMPP_MT_ROUTE_RESP CommandId = 0x00000010 - 10 + iota, 0x80000010 - 10 + iota
	CMPP_MO_ROUTE, CMPP_MO_ROUTE_RESP
	CMPP_GET_MT_ROUTE, CMPP_GET_MT_ROUTE_RESP
	CMPP_MT_ROUTE_UPDATE, CMPP_MT_ROUTE_UPDATE_RESP
	CMPP_MO_ROUTE_UPDATE, CMPP_MO_ROUTE_UPDATE_RESP
	CMPP_PUSH_MT_ROUTE_UPDATE, CMPP_PUSH_MT_ROUTE_UPDATE_RESP
	CMPP_PUSH_MO_ROUTE_UPDATE, CMPP_PUSH_MO_ROUTE_UPDATE_RESP
	CMPP_GET_MO_ROUTE, CMPP_GET_MO_ROUTE_RESP
	CMPP_REQUEST_MAX, CMPP_RESPONSE_MAX
)

func (id CommandId) String() string {
	if id <= CMPP_FWD && id > CMPP_REQUEST_MIN {
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
	} else if id >= CMPP_MT_ROUTE && id < CMPP_REQUEST_MAX {
		return []string{
			"CMPP_MT_ROUTE",
			"CMPP_MO_ROUTE",
			"CMPP_GET_MT_ROUTE",
			"CMPP_MT_ROUTE_UPDATE",
			"CMPP_MO_ROUTE_UPDATE",
			"CMPP_PUSH_MT_ROUTE_UPDATE",
			"CMPP_PUSH_MO_ROUTE_UPDATE",
			"CMPP_GET_MO_ROUTE",
		}[id-0x00000010]
	}

	if id <= CMPP_FWD_RESP && id > CMPP_RESPONSE_MIN {
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
	} else if id >= CMPP_MT_ROUTE_RESP && id < CMPP_RESPONSE_MAX {
		return []string{
			"CMPP_MT_ROUTE_RESP",
			"CMPP_MO_ROUTE_RESP",
			"CMPP_GET_MT_ROUTE_RESP",
			"CMPP_MT_ROUTE_UPDATE_RESP",
			"CMPP_MO_ROUTE_UPDATE_RESP",
			"CMPP_PUSH_MT_ROUTE_UPDATE_RESP",
			"CMPP_PUSH_MO_ROUTE_UPDATE_RESP",
			"CMPP_GET_MO_ROUTE_RESP",
		}[id-0x80000010]
	}
	return "unknown"
}
