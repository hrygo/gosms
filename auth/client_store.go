package auth

import (
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/hrygo/log"
	"gopkg.in/yaml.v3"

	bs "github.com/hrygo/gosmsn/bootstrap"
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
	}
	// TODO 其他类型存储的判断

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

// 公共代码结束 _________________________________
// 以下为基于Yaml文件的实现的实现 _________________________________

type YamlStore storage

func (y *YamlStore) Load() {
	dir := bs.ConfigYml.GetString("AuthClient.YamlFilePath")
	if "" == dir {
		dir = "config/clients/"
	}
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	dir = bs.BasePath + dir
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("Config file init error: %v", err)
	}

	for _, f := range fs {
		f := f
		if f.IsDir() {
			continue
		}
		// e.g. cmpp_123456
		var key string
		if strings.HasSuffix(f.Name(), ".yaml") {
			key = f.Name()[:len(f.Name())-5]
		}
		if strings.HasSuffix(f.Name(), ".yml") {
			key = f.Name()[:len(f.Name())-4]
		}
		if "" == key {
			continue
		}
		data, err := ioutil.ReadFile(dir + f.Name())
		if err != nil {
			log.Fatalf("Config file init error: %v", err)
		}

		unmarshal(data, strings.ToLower(key), y)
	}
}

func (y *YamlStore) FindByCid(isp string, cid string) *Client {
	y.Lock()
	defer y.Unlock()
	client := y.cache[strings.ToLower(isp)+"_"+cid]
	return client
}

func unmarshal(data []byte, id string, y *YamlStore) {
	y.Lock()
	defer y.Unlock()

	cli := &Client{}
	err := yaml.Unmarshal(data, cli)
	if err != nil {
		log.Errorf("Config file init error: %v", err)
	} else {
		y.cache[id] = cli
	}
}
