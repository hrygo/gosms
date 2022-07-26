package server

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/hrygo/log"

	bs "github.com/hrygo/gosmsn/bootstrap"
	"github.com/hrygo/gosmsn/codec"
)

const (
	SrvName    = "server" // 会话标识
	Sid        = "sid"    // 会话标识
	RemoteAddr = "remote" // 发起方地址

	LogKeyErr              = "error"
	LogKeyThreshold        = "threshold"
	LogKeyActiveConns      = "conns_current"
	LogKeyConnsThreshold   = "conns_threshold"
	LogKeyReceiveWindowLen = "r_window_len"
	LogKeyReceiveWindowCap = "r_window_cap"
	LogKeyPacket           = "packet"
	LogKeySessionId        = "sid"
)

type Direction string

var RC Direction = "<<<" // 交易请求方向 接收 Remote > Local
var SD Direction = ">>>" // 交易请求方向 发送 Local  > Remote

func SSR(s *session, remote net.Addr, cap ...int) (ret []log.Field) {
	if len(cap) == 0 || cap[0] < 8 {
		ret = make([]log.Field, 0, 16)
	} else {
		ret = make([]log.Field, 0, cap[0])
	}
	if s != nil {
		ret = append(ret, log.String(SrvName, s.serverName), log.Uint64(Sid, s.Id()))
	}
	if remote != nil {
		ret = append(ret, log.String(RemoteAddr, remote.String()))
	}
	return
}

func JoinLog(into []log.Field, fields ...log.Field) []log.Field {
	return append(into, fields...)
}

func CCWW(s *Server) []log.Field {
	return []log.Field{
		log.Int(LogKeyActiveConns, s.engine.CountConnections()),
		log.Int(LogKeyConnsThreshold, bs.ConfigYml.GetInt("Server."+s.name+".MaxConnections")),
		log.Int(LogKeyReceiveWindowLen, len(s.window)),
		log.Int(LogKeyReceiveWindowCap, cap(s.window)),
	}
}

func Packet2HexLogStr(pack []byte) log.Field {
	return log.String(LogKeyPacket, hex.EncodeToString(pack))
}

func Packet2LogStr(pdu codec.Pdu) log.Field {
	return log.String(LogKeyPacket, fmt.Sprintf("%v", pdu))
}

// Operation 定义操作类型
type Operation byte

const (
	operation = "op" // 操作类型

	FlowControl Operation = iota // 操作类型枚举
	ConnectionClose
	ActiveTest
)

func (op Operation) String() string {
	return []string{
		"flow_control",
		"connection_close",
		"active_test",
	}[op-1]
}

func (op Operation) Field() log.Field {
	return log.String(operation, op.String())
}

// Reason 操作对应的具体原因
type Reason byte

const (
	reason = "reason" // 操作原因

	TotalConnectionsThresholdReached Reason = iota // 操作原因类型枚举
	TotalReceiveWindowsThresholdReached
	NoEffectiveActionTimeThresholdReached
	NoneResponseActiveTestCountThresholdReached
)

func (op Reason) String() string {
	return []string{
		"total_connections_threshold_reached",
		"total_receive_window_threshold_reached",
		"no effective action more than 5 minutes",
		"none response active test 3 times",
	}[op-1]
}

func (op Reason) Field() log.Field {
	return log.String(reason, op.String())
}

func ErrorField(err error) log.Field {
	if err == nil {
		return log.String(LogKeyErr, "<nil>")
	}
	return log.String(LogKeyErr, err.Error())
}
