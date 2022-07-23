package my_log

import (
	"github.com/hrygo/log"
	"go.uber.org/zap"

	bs "github.com/hrygo/gosmsn/bootstrap"
)

// New 创建一个自定义日志：
// name 对应 config.yaml 配置文件中的配置中间名；
// logLevel 可输出的最小日志级别；
// opts 为日志增加可选项。
func New(name string, logLevel log.Level, opts ...log.Option) *zap.Logger {
	var top = []log.TeeOption{{
		Filename:      bs.BasePath + bs.ConfigYml.GetString("Logs."+name+".LogName"),
		TextFormat:    bs.ConfigYml.GetString("Logs." + name + ".TextFormat"),
		TimePrecision: bs.ConfigYml.GetString("Logs." + name + ".TimePrecision"),
		Ropt: log.RotateOptions{
			MaxSize:    bs.ConfigYml.GetInt("Logs." + name + ".MaxSize"),
			MaxAge:     bs.ConfigYml.GetInt("Logs." + name + ".MaxBackups"),
			MaxBackups: bs.ConfigYml.GetInt("Logs." + name + ".MaxAge"),
			Compress:   bs.ConfigYml.GetBool("Logs." + name + ".Compress"),
		},
		LvlEnableFunc: func(lvl log.Level) bool {
			return lvl >= logLevel
		},
	}}
	return log.NewTeeWithRotate(top, opts...)
}
