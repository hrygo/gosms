package test_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosms/auth"
	sms "github.com/hrygo/gosms/smc_client"
)

func init() {
	// 启动记录数据库的程序
	if sms.ConfigYml.GetString("Mongo.URI") != "" {
		sms.PersistenceSmsJournal()
	} else {
		sms.StartCacheExpireTicker(nil)
	}
	auth.Cache = auth.New(sms.ConfigYml)
}

func TestSend(t *testing.T) {
	i := 100
	for i > 0 {
		queryId := sms.Send("hello world", "13800001111")
		assert.True(t, len(sms.Query(queryId)) > 0)
		i--
	}
	i = 100
	for i > 0 {
		queryId := sms.Send("hello world", "13300001111")
		assert.True(t, len(sms.Query(queryId)) > 0)
		i--
	}
	time.Sleep(time.Minute)
}
