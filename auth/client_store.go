package auth

import (
	"sync"
	"time"

	"github.com/hrygo/log"

	bs "github.com/hrygo/gosms/bootstrap"
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
	cache map[string]*Client
}

func init() {
	st := bs.ConfigYml.GetString("AuthClient.StoreType")
	if "" == st || "yaml" == st || "yml" == st {
		Cache = &YamlStore{cache: make(map[string]*Client)}
	} else if "mongo" == st {
		db.InitDB(bs.ConfigYml, "AuthClient.Mongo")
		Cache = &MongoStore{cache: make(map[string]*Client)}
	}
	// 初次加载存储
	Cache.Load()
	// 启动定时器，定时加载存储
	startTicker(Cache)
}

func startTicker(s Store) {
	go func(s Store) {
		d := bs.ConfigYml.GetDuration("AuthClient.ReloadTicker")
		if d == 0 {
			log.Warn("Client.ReloadTicker configuration missing, ticker not started.")
			return
		}
		ticker := time.NewTicker(d)
		defer ticker.Stop()

		for {
			<-ticker.C
			log.Warn("auth cache reload.")
			s.Load()
		}
	}(s)
}
