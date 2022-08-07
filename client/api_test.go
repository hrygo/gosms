package sms_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hrygo/gosmsn/bootstrap"
	sms "github.com/hrygo/gosmsn/client"
)

var _ = bootstrap.BasePath

func init() {
	sms.PersistenceSmsJournal()
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
