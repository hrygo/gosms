package sms

import (
	"os"
	"strings"
	"time"

	"github.com/hrygo/log"
	"github.com/hrygo/yaml_config"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"
	"go.uber.org/zap"

	"github.com/hrygo/gosms/codec"
	"github.com/hrygo/gosms/codec/sgip"
	"github.com/hrygo/gosms/event_manager"
	"github.com/hrygo/gosms/utils"
	"github.com/hrygo/gosms/utils/snowflake"
)

const (
	ProjectName       = "msc_client"
	DefaultConfigPath = "config/"
)

var (
	BasePath  = ""
	ConfigYml yaml_config.YmlConfig
	statChan  = make(chan struct{})
	pool      *goroutine.Pool
)

func init() {
	log.Info(ProjectName + "\tstart initialization ...")
	defer log.Info(ProjectName + "\tfinished initialization.")

	// 0. 设置初始化路径
	setBasePath(ProjectName)

	// 1. 读取配置文件(默认配置文件 config.yaml)
	ConfigYml = yaml_config.CreateYamlFactory(DefaultConfigPath, "", ProjectName)
	ConfigYml.ConfigFileChangeListen()

	// 2. 初始化日志框架
	debug := ConfigYml.GetBool("AppDebug")
	if !debug {
		logInit()
	}

	// 3. 初始化序号器
	SeqInit()

	// 4. 初始化线程池
	poolInit()
}

func StatChan() <-chan struct{} {
	return statChan
}

func setBasePath(project string) {
	if curPath, err := os.Getwd(); err == nil {
		log.Debugf("os.Getwd: %s", curPath)
		for _, arg := range os.Args {
			log.Debugf("os.Args: %s", arg)
		}
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
			Level: log.Level(ConfigYml.GetInt("Logs.Default.Level")),
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
			Level: log.Level(ConfigYml.GetInt("Logs.Default.Level")),
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
	var b64worker = ConfigYml.GetInt64("Snowflake.B64.worker")
	var b32dc = ConfigYml.GetInt64("Snowflake.B64.DC")
	var b32worker = ConfigYml.GetInt64("Snowflake.B64.worker")
	var bcd = ConfigYml.GetString("Snowflake.BCD")
	var node = ConfigYml.GetInt32("Snowflake.SGIP")

	codec.B64Seq = snowflake.NewSnowflake(b64dc, b64worker)
	codec.B32Seq = utils.NewCycleSequence(int32(b32dc), int32(b32worker))
	codec.BcdSeq = utils.NewBcdSequence(bcd)
	sgip.NewSequencer(uint32(node), uint32(8*b32dc+b32worker))
}

func poolInit() {
	poolSize := ConfigYml.GetInt("PoolSize")
	if poolSize < 10 {
		poolSize = 10
	}
	var options = ants.Options{
		ExpiryDuration: time.Minute, // 1 分钟内不被使用的worker会被清除
		Nonblocking:    false,       // 如果为true,worker池满了后提交任务会直接返回nil
		PreAlloc:       false,
		PanicHandler: func(e interface{}) {
			log.Errorf("%v", e)
		},
	}
	pool, _ = ants.NewPool(poolSize, ants.WithOptions(options))
	event_manager.RegisterShutdownHooker("cache_pool_release", func(args ...any) {
		pool.Release()
	})
}
