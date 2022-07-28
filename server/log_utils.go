package server

import (
	"encoding/hex"

	"github.com/hrygo/log"
)

const (
	Sid                   = "sid"    // 会话标识
	SrvName               = "server" // 会话标识
	RemoteAddr            = "remote" // 发起方地址
	LogKeyErr             = "error"  // 错误信息
	LogKeyPacket          = "packet" // 数据包
	_                     = "【全 局】计数器"
	LogKeyConnsCurrent    = "g_conn_cur"  // 【全局】当前连接数
	LogKeyConnsCap        = "g_conn_cap"  // 【全局】最大连接数
	LogKeySwCur           = "g_sw_cur"    // 【全局】接收消息滑动窗口当前大小
	LogKeySwCap           = "g_sw_cap"    // 【全局】接收消息滑动窗口最大值
	LogKeyPoolFree        = "g_pool_free" // 【全局】goroutine pool 当前使用数
	LogKeyPoolCap         = "g_pool_cap"  // 【全局】goroutine pool 最大容量
	_                     = "【会话级】计数器"
	LogKeySessionConnsCap = "s_conn_cap"  // 【会话级】最大连接数
	LogKeySessionSwCur    = "s_sw_cur"    // 【会话级】接收消息滑动窗口当前大小
	LogKeySessionSwCap    = "s_sw_cap"    // 【会话级】接收消息滑动窗口最大值
	LogKeySessionPoolFree = "s_pool_free" // 【会话级】goroutine pool 当前使用数
	LogKeySessionPoolCap  = "s_pool_cap"  // 【会话级】goroutine pool 最大容量
	LogKeyCounterMt       = "s_mt_num"    // 【会话级】下行短信计数
	LogKeyCounterDlv      = "s_dlv_num"   // 【会话级】上行短信计数
	LogKeyCounterRpt      = "s_rpt_num"   // 【会话级】状态报告计数
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
)

func (op Operation) String() string {
	return []string{
		"flow_control",
		"connection_close",
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
