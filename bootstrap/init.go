package bootstrap

import (
	"os"
	"strings"

	"github.com/hrygo/log"
	"github.com/hrygo/yaml_config"
	"go.uber.org/zap"

	"github.com/hrygo/gosmsn/codec"
	"github.com/hrygo/gosmsn/my_errors"
	"github.com/hrygo/gosmsn/utils"
	"github.com/hrygo/gosmsn/utils/snowflake"
)

const (
	ProjectName       = "gosmsn"
	DefaultConfigPath = "config/"
	DefaultConfigFile = DefaultConfigPath + "config.yaml"
)

var (
	BasePath  = ""
	ConfigYml yaml_config.YmlConfig
	statChan  = make(chan struct{})
)

func init() {
	log.Info(ProjectName + "\tstart initialization ...")
	defer log.Info(ProjectName + "\tfinished initialization.")

	// 0. 设置初始化路径
	setBasePath(ProjectName)

	// 1. 检查配置文件是否存在
	checkRequiredFolders()

	// 2. 读取配置文件(默认配置文件 config.yaml)
	ConfigYml = yaml_config.CreateYamlFactory(DefaultConfigPath, "", ProjectName)
	ConfigYml.ConfigFileChangeListen()

	// 3. 初始化日志框架
	debug := ConfigYml.GetBool("AppDebug")
	if !debug {
		logInit()
	}

	// 4. 初始化序号器
	SeqInit()
}

func StatChan() <-chan struct{} {
	return statChan
}

func setBasePath(project string) {
	if curPath, err := os.Getwd(); err == nil {
		// 路径进行处理，兼容单元测试程序程序启动时的奇怪路径
		pl, cl := len(project), len(curPath)
		if pl != 0 && cl > pl && len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "-test") {
			i := strings.Index(curPath, project)
			if i > 0 {
				BasePath = curPath[:i] + project
			}
		} else {
			BasePath = curPath
		}
		BasePath += "/"
	} else {
		log.Fatal("Running directory has no permission!")
	}
}

// 检查项目必须的非编译目录是否存在，避免编译后调用的时候缺失相关目录
func checkRequiredFolders() {
	// 检查配置文件 DefaultConfigFile 是否存在
	if _, err := os.Stat(BasePath + DefaultConfigFile); err != nil {
		log.Fatal(my_errors.ErrorsConfigYamlNotExists, log.String("error", err.Error()))
	}
}

func logInit() {
	var tops = []log.TeeOption{
		{
			Filename:      BasePath + ConfigYml.GetString("Logs.Default.LogName"),
			TextFormat:    ConfigYml.GetString("Logs.Default.TextFormat"),
			TimePrecision: ConfigYml.GetString("Logs.Default.TimePrecision"),
			Ropt: log.RotateOptions{
				MaxSize:    ConfigYml.GetInt("Logs.Default.MaxSize"),
				MaxAge:     ConfigYml.GetInt("Logs.Default.MaxAge"),
				MaxBackups: ConfigYml.GetInt("Logs.Default.MaxBackups"),
				Compress:   ConfigYml.GetBool("Logs.Default.Compress"),
			},
			LvlEnableFunc: func(lvl log.Level) bool {
				return lvl >= log.Level(ConfigYml.GetInt("Logs.Default.Level"))
			},
		},
		{
			Filename:      BasePath + ConfigYml.GetString("Logs.Error.LogName"),
			TextFormat:    ConfigYml.GetString("Logs.Error.TextFormat"),
			TimePrecision: ConfigYml.GetString("Logs.Error.TimePrecision"),
			Ropt: log.RotateOptions{
				MaxSize:    ConfigYml.GetInt("Logs.Error.MaxSize"),
				MaxAge:     ConfigYml.GetInt("Logs.Error.MaxAge"),
				MaxBackups: ConfigYml.GetInt("Logs.Error.MaxBackups"),
				Compress:   ConfigYml.GetBool("Logs.Error.Compress"),
			},
			LvlEnableFunc: func(lvl log.Level) bool {
				return lvl >= log.Level(ConfigYml.GetInt("Logs.Error.Level"))
			},
		},
	}

	logger := log.NewTeeWithRotate(
		tops,
		zap.AddStacktrace(zap.ErrorLevel),
		log.WithCaller(true),
		// TODO 可添加其他日志Option，如 zap.Hooks
	)
	log.ResetDefault(logger)
}

func SeqInit() {
	var b64dc = ConfigYml.GetInt64("Snowflake.B64.DC")
	var b64worker = ConfigYml.GetInt64("Snowflake.B64.Worker")
	var b32dc = ConfigYml.GetInt64("Snowflake.B64.DC")
	var b23worker = ConfigYml.GetInt64("Snowflake.B64.Worker")
	var bcd = ConfigYml.GetString("Snowflake.BCD")

	codec.B64Seq = snowflake.NewSnowflake(b64dc, b64worker)
	codec.B32Seq = utils.NewCycleSequence(int32(b32dc), int32(b23worker))
	codec.BcdSeq = utils.NewBcdSequence(bcd)
}
