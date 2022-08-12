package auth

import (
	"github.com/hrygo/yaml_config"
)

var ConfigYml yaml_config.YmlConfig

func init() {
	// 2. 读取配置文件(默认配置文件 config.yaml)
	ConfigYml = yaml_config.CreateYamlFactory("", "", "auth_test")
	ConfigYml.ConfigFileChangeListen()
}
