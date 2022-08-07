package auth

import (
	"time"
)

type Client struct {
	ISP             string        `yaml:"isp"`               // 即运营商标识 cmpp、sgip、smgp
	ClientId        string        `yaml:"client-id"`         // 即SourceAddr
	SharedSecret    string        `yaml:"shared-secret"`     // 通讯密码
	Version         byte          `yaml:"version"`           // 见CMPP协议，48表示3.0 即 0x30 = 0011 0000
	NeedReport      byte          `yaml:"need-report"`       // 是否需状态报告
	SmsDisplayNo    string        `yaml:"sms-display-no"`    // 发送号码（后面可拼接子码）
	ServiceId       string        `yaml:"service-id"`        // 运营商分配的服务ID
	DefaultMsgLevel byte          `yaml:"default-msg-level"` // 默认短信优先级 （范围1-9）
	FeeUserType     byte          `yaml:"fee-user-type"`     // 费用相关
	FeeTerminalType byte          `yaml:"fee-terminal-type"` // 费用相关
	FeeTerminalId   string        `yaml:"fee-terminal-id"`   // 费用相关
	FeeType         string        `yaml:"fee-type"`          // 费用相关
	FeeCode         string        `yaml:"fee-code"`          // 费用相关
	FixedFee        string        `yaml:"fixed-fee"`         // 费用相关
	LinkId          string        `yaml:"link-id"`           // 点播业务相关
	MaxConns        int           `yaml:"max-conns"`         // 最大连接数
	MtWindowSize    int           `yaml:"mt-window-size"`    // 接收窗口大小,服务端分配
	MtValidDuration time.Duration `yaml:"mt-valid-duration"` // 短信默认有效期，超过下面配置时长后，如果消息未发送，则不再发送
	Throughput      int           `yaml:"throughput"`        // 系统最大吞吐,单位tps
}
