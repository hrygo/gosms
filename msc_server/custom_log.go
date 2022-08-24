package msc

import (
	"github.com/hrygo/log"
	"go.uber.org/zap"
)

// New 创建一个自定义日志：
// name 对应 config.yaml 配置文件中的配置中间名（如果找不到，会采用默认配置）；
// opts 为日志增加可选项。
func New(name string, opts ...log.Option) *zap.Logger {
	fileName := ConfigYml.GetString("Logs." + name + ".LogName")
	if fileName == "" {
		fileName = "logs/" + name + ".log"
	}
	textFormat := ConfigYml.GetString("Logs." + name + ".TextFormat")
	if textFormat == "" {
		textFormat = "json"
	}
	timePrecision := ConfigYml.GetString("Logs." + name + ".TimePrecision")
	if timePrecision == "" {
		timePrecision = "millisecond"
	}
	maxSize := ConfigYml.GetInt("Logs." + name + ".MaxSize")
	if maxSize == 0 {
		maxSize = 10
	}
	backups := ConfigYml.GetInt("Logs." + name + ".MaxBackups")
	if backups == 0 {
		backups = 100
	}
	maxAge := ConfigYml.GetInt("Logs." + name + ".MaxAge")
	if maxAge == 0 {
		maxAge = 14
	}
	compress := ConfigYml.GetBool("Logs." + name + ".Compress")
	level := log.Level(ConfigYml.GetInt("Logs" + name + "Level"))

	var top = []log.TeeOption{{
		Filename:      BasePath + fileName,
		TextFormat:    textFormat,
		TimePrecision: timePrecision,
		Ropt: log.RotateOptions{
			MaxSize:    maxSize,
			MaxAge:     maxAge,
			MaxBackups: backups,
			Compress:   compress,
		},
		Level: level,
	}}
	return log.NewTeeWithRotate(top, opts...)
}
