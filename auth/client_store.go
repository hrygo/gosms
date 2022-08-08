package auth

import (
	"sync"
	"time"

	"github.com/hrygo/log"
	"github.com/hrygo/yaml_config"

	db "github.com/hrygo/gosms/databse"
)

type Store interface {
	// Load 从存储加载客户端配置信息
	Load()
	// FindByCid 根据客户端ID获取指定客户端配置信息:
	// isp 运营商，用协议名称表示 CMPP、SGIP、SMGP
	FindByCid(isp string, cid string) *Client
	// 采用定时器，定时刷新配置
}

// Cache 从存储加载的客户端缓存数据
var Cache Store

type storage struct {
	sync.Mutex
	cache  map[string]*Client
	config yaml_config.YmlConfig
}

func New(c yaml_config.YmlConfig) (cache Store) {
	st := c.GetString("AuthClient.StoreType")
	if "" == st || "yaml" == st || "yml" == st {
		cache = &YamlStore{
			config: c,
			cache:  make(map[string]*Client),
		}
	} else if "mongo" == st {
		db.InitDB(c, "AuthClient.Mongo")
		cache = &MongoStore{
			config: c,
			cache:  make(map[string]*Client),
		}
	}
	// 初次加载存储
	cache.Load()
	// 启动定时器，定时加载存储
	startTicker(c, cache)
	return
}

func startTicker(c yaml_config.YmlConfig, s Store) {
	go func() {
		d := c.GetDuration("AuthClient.ReloadTicker")
		if d == 0 {
			log.Warn("Client.ReloadTicker configuration missing, use default value: 5 minutes.")
			d = 5 * time.Minute
		}
		ticker := time.NewTicker(d)
		defer ticker.Stop()

		for {
			<-ticker.C
			log.Warn("Auth cache reload!")
			s.Load()
		}
	}()
}
