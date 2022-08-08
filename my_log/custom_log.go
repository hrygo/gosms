package my_log

import (
	"github.com/hrygo/log"
	"go.uber.org/zap"

	bs "github.com/hrygo/gosms/bootstrap"
)

// New 创建一个自定义日志：
// name 对应 config.yaml 配置文件中的配置中间名（如果找不到，会采用默认配置）；
// opts 为日志增加可选项。
func New(name string, opts ...log.Option) *zap.Logger {
	fileName := bs.ConfigYml.GetString("Logs." + name + ".LogName")
	if fileName == "" {
		fileName = "logs/" + name + ".log"
	}
	textFormat := bs.ConfigYml.GetString("Logs." + name + ".TextFormat")
	if textFormat == "" {
		textFormat = "json"
	}
	timePrecision := bs.ConfigYml.GetString("Logs." + name + ".TimePrecision")
	if timePrecision == "" {
		timePrecision = "millisecond"
	}
	maxSize := bs.ConfigYml.GetInt("Logs." + name + ".MaxSize")
	if maxSize == 0 {
		maxSize = 10
	}
	backups := bs.ConfigYml.GetInt("Logs." + name + ".MaxBackups")
	if backups == 0 {
		backups = 100
	}
	maxAge := bs.ConfigYml.GetInt("Logs." + name + ".MaxAge")
	if maxAge == 0 {
		maxAge = 14
	}
	compress := bs.ConfigYml.GetBool("Logs." + name + ".Compress")
	level := log.Level(bs.ConfigYml.GetInt("Logs" + name + "Level"))

	var top = []log.TeeOption{{
		Filename:      bs.BasePath + fileName,
		TextFormat:    textFormat,
		TimePrecision: timePrecision,
		Ropt: log.RotateOptions{
			MaxSize:    maxSize,
			MaxAge:     maxAge,
			MaxBackups: backups,
			Compress:   compress,
		},
		LvlEnableFunc: func(lvl log.Level) bool {
			return lvl >= level
		},
	}}
	return log.NewTeeWithRotate(top, opts...)
}
