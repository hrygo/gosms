package auth

import (
	"time"
)

type Client struct {
	ISP             string        `yaml:"isp"               json:"isp"`             // 即运营商标识 cmpp、sgip、smgp
	ClientId        string        `yaml:"client-id"         json:"clientId"`        // 即SourceAddr
	SharedSecret    string        `yaml:"shared-secret"     json:"sharedSecret"`    // 通讯密码
	Version         byte          `yaml:"version"           json:"version"`         // 见CMPP协议，48表示3.0 即 0x30 = 0011 0000
	NeedReport      byte          `yaml:"need-report"       json:"needReport"`      // 是否需状态报告
	SmsDisplayNo    string        `yaml:"sms-display-no"    json:"smsDisplayNo"`    // 发送号码（后面可拼接子码）
	ServiceId       string        `yaml:"service-id"        json:"serviceId"`       // 运营商分配的服务ID
	DefaultMsgLevel byte          `yaml:"default-msg-level" json:"DefaultMsgLevel"` // 默认短信优先级 （范围1-9）
	FeeUserType     byte          `yaml:"fee-user-type"     json:"feeUserType"`     // 费用相关
	FeeTerminalType byte          `yaml:"fee-terminal-type" json:"FeeTerminalType"` // 费用相关
	FeeTerminalId   string        `yaml:"fee-terminal-id"   json:"feeTerminalId"`   // 费用相关
	FeeType         string        `yaml:"fee-type"          json:"feeType"`         // 费用相关
	FeeCode         string        `yaml:"fee-code"          json:"feeCode"`         // 费用相关
	FixedFee        string        `yaml:"fixed-fee"         json:"fixedFee"`        // 费用相关
	LinkId          string        `yaml:"link-id"           json:"LinkId"`          // 点播业务相关
	MtValidDuration time.Duration `yaml:"mt-valid-duration" json:"mtValidDuration"` // 短信默认有效期，超过下面配置时长后，如果消息未发送，则不再发送
	MaxConns        int           `yaml:"max-conns"         json:"maxConns"`        // 最大连接数
	MtWindowSize    int           `yaml:"mt-window-size"    json:"mtWindowSize"`    // 接收窗口大小,服务端分配
	Throughput      int           `yaml:"throughput"        json:"throughput"`      // 系统最大吞吐,单位tps
}
