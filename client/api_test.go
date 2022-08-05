package sms_test

import (
	"testing"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/bootstrap"
	sms "github.com/hrygo/gosmsn/client"
	"github.com/hrygo/gosmsn/client/session"
)

var _ = bootstrap.BasePath

func init() {
	sms.StartCacheExpireTicker(func(results []*session.Result) {
		log.Infof("[Persistence] Save %d send results to db.", len(results))
		for _, result := range results {
			log.Infof("[Persistence] %v", result)
		}
	})
}

func TestSend(t *testing.T) {
	i := 100
	for i > 0 {
		sms.Send("hello world", "13800001111")
		i--
	}
	i = 100
	for i > 0 {
		sms.Send("hello world", "13300001111")
		i--
	}
	time.Sleep(time.Minute)
}
