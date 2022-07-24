package server

import (
	"encoding/hex"

	"github.com/hrygo/log"
)

const (
	LogKeyErr            = "error"
	LogKeyThreshold      = "threshold"
	LogKeyActiveConns    = "active_conns"
	LogKeyCurrentWinSize = "cur_window_size"
	LogKeyPacket         = "packet"
	LogKeySessionId      = "sid"
)

func Packet2HexLogStr(pack []byte) log.Field {
	return log.String(LogKeyPacket, hex.EncodeToString(pack))
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
)

func (op Reason) String() string {
	return []string{
		"total_connections_threshold_reached",
		"total_receive_window_threshold_reached",
		"no effective action more than 5 minutes",
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
