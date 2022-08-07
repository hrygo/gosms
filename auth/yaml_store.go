package auth

import (
	"io/ioutil"
	"strings"

	"github.com/hrygo/log"
	"gopkg.in/yaml.v3"

	bs "github.com/hrygo/gosmsn/bootstrap"
)

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
		cli.ISP = id[0:4]
		y.cache[id] = cli
	}
}
