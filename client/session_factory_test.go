package sms_test

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"testing"
	"time"

	"github.com/hrygo/log"
	"github.com/stretchr/testify/assert"

	bs "github.com/hrygo/gosmsn/bootstrap"
	sms "github.com/hrygo/gosmsn/client"
)

func TestSessionFactory(t *testing.T) {
	pprofDebug()

	var fa = sms.CreateSessionFactory("CMPP")
	var fa1 = sms.CreateSessionFactory("CMPP")
	assert.True(t, fa != nil)
	assert.True(t, fa1 != nil)
	assert.True(t, fa == fa1)
	fa1 = sms.CreateSessionFactory("SMGP")
	assert.True(t, fa != fa1)

	sc := fa.PeekSession()
	assert.True(t, sc != nil)
	sc.AddCounter()
	sc2 := fa.PeekSession()
	for sc == sc2 {
		sc.AddCounter()
		time.Sleep(time.Millisecond)
		sc2 = fa.PeekSession()
	}
	assert.True(t, sc != sc2)

	sc.Close()
	sc2.Close()

	time.Sleep(3600 * time.Second)
}

func TestPanic(t *testing.T) {
	defer func() {
		err := recover()
		assert.True(t, err != nil)
	}()

	var fa = sms.CreateSessionFactory("ABCD")
	assert.True(t, fa != nil)

}

// 开启pprof，监听请求
func pprofDebug() {
	if bs.ConfigYml.GetBool("Server.Pprof.Enable") {
		go func() {
			var pprof = 10099
			log.Warnf("debug pprof on http://localhost:%d/debug/pprof/", pprof)
			if err := http.ListenAndServe(fmt.Sprintf(":%d", pprof), nil); err != nil {
				log.Fatalf("start pprof failed on %s", pprof)
			}
		}()
	}
}

func TestSelectSession(t *testing.T) {
	session := sms.SelectSession("13800001111")

	assert.True(t, session != nil)
}
