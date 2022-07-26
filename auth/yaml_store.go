package auth

import (
	"io/ioutil"
	"strings"

	"github.com/hrygo/log"
	"gopkg.in/yaml.v3"
)

type YamlStore storage

func (y *YamlStore) Load() {
	dir := y.Config.GetString("AuthClient.YamlFilePath")
	if "" == dir {
		dir = "Config/yml_store/"
	}
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	dir = y.Config.BasePath() + dir
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
	client := y.Cache[strings.ToLower(isp)+"_"+cid]
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
		cli.ISP = id[0:4]
		y.Cache[id] = cli
	}
}
