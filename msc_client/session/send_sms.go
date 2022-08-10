package session

import (
	"sync"
	"time"

	"github.com/hrygo/gosms/codec"
)

type Result struct {
	QueryId      int64     `json:"QueryId"`      // 给客户端用的查询编号
	Phone        string    `json:"phone"`        // 手机号
	SequenceId   uint32    `json:"sequenceId"`   // 消息发送的标识
	Result       uint32    `json:"result"`       // 消息发送的网关响应码
	MsgId        string    `json:"msgId"`        // 消息的msgId用于关联状态报告
	Report       string    `json:"report"`       // DELIVRD 等7直接状态码
	SendTime     time.Time `json:"sendTime"`     // 发送时间
	ResponseTime time.Time `json:"responseTime"` // 网关响应时间
	ReportTime   time.Time `json:"reportTime"`   // 状态报告时间
}

// SequenceIdResultCacheMap 临时存储短信发送的返回结果数据，Key为requestId,value为*Result，后续采用数据库存储
var SequenceIdResultCacheMap sync.Map

// MsgIdResultCacheMap 临时存储短信发送的返回结果数据，Key为msgId,value为*Result，后续采用数据库存储
var MsgIdResultCacheMap sync.Map

// Send 发送短信
func (s *Session) Send(phone string, message string, options ...codec.OptionFunc) []any {
	switch s.serverName {
	case CMPP:
		return s.sendByCmpp(phone, message, options...)
	case SMGP:
		return s.sendBySmgp(phone, message, options...)
	case SGIP:
		return s.sendBySgip(phone, message, options...)
	}
	return nil
}
