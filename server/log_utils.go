package server

import (
	"encoding/hex"

	"github.com/hrygo/log"
)

const (
	Sid                    = "sid"    // 会话标识
	SrvName                = "server" // 会话标识
	RemoteAddr             = "remote" // 发起方地址
	LogKeyErr              = "error"
	LogKeyActiveConns      = "conns"
	LogKeyConnsThreshold   = "conns_max"
	LogKeyReceiveWindowLen = "sw_len"
	LogKeyReceiveWindowCap = "sw_max"
	LogKeyPacket           = "packet"
	LogKeyCounterMt        = "c_mt"
	LogKeyCounterDlv       = "c_dlv"
	LogKeyCounterRpt       = "c_rpt"
)

type Direction string

var RC Direction = "<<<" // 交易请求方向 接收 Remote > Local
var SD Direction = ">>>" // 交易请求方向 发送 Local  > Remote

func FlatMapLog(fields ...[]log.Field) []log.Field {
	if len(fields) == 0 {
		return make([]log.Field, 0)
	} else {
		ret := fields[0]
		for i := 1; i < len(fields); i++ {
			ret = append(ret, fields[i]...)
		}
		return ret
	}
}

func Packet2HexLogStr(pack []byte) log.Field {
	return log.String(LogKeyPacket, hex.EncodeToString(pack))
}

// Operation 操作类型定义
type Operation byte

const (
	operation               = "op" // 操作类型
	OpFlowControl Operation = iota // 操作类型枚举
	OpConnectionClose
	OpActiveTest
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

func SErrField(err string) log.Field {
	return log.String(LogKeyErr, err)
}

func ErrorField(err error) log.Field {
	if err != nil {
		return log.String(LogKeyErr, err.Error())
	} else {
		return log.String(LogKeyErr, "<nil>")
	}
}
