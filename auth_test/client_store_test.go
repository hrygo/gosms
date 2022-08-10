package auth

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/hrygo/gosms/auth"
)

func TestYamlStore_FindByCid(t *testing.T) {
	auth.Cache = auth.New(ConfigYml)
	c := auth.Cache.FindByCid("SMGP", "12345678")
	assert.True(t, c != nil)
	t.Logf("%#+v", c)
	jsn, err := json.MarshalIndent(c, "", "    ")
	assert.True(t, err == nil)
	t.Log("\n" + string(jsn))

	assert.True(t, c.ClientId == "12345678")
	assert.True(t, c.Version == 0x30)
}

type YamlCase struct {
	AbCd string  `yaml:"ab-cd"`
	Fn   float64 `yaml:"fn"`
	fn   float64 `yaml:"fn"`
	de   int     `yaml:"de"`
}

func (y *YamlCase) SetFn(f float64) {
	y.fn = f
}

var yamlStr = `ab-cd: hello world
de: 128
fn: 3.14`

// 测试证明只有以导出属性（大写开头）才可以在Unmarshal时被赋值
func TestYamlUnmarshal(t *testing.T) {
	y := &YamlCase{}
	_ = yaml.Unmarshal([]byte(yamlStr), y)

	assert.True(t, y.AbCd == "hello world")
	assert.True(t, y.Fn == 3.14)
	assert.True(t, y.fn == 0)
	assert.True(t, y.de == 0)
}
