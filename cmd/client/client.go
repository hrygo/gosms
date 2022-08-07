package main

import (
	"flag"
	"os"
	"time"

	"github.com/hrygo/log"

	bs "github.com/hrygo/gosmsn/bootstrap"
	sms "github.com/hrygo/gosmsn/client"
	"github.com/hrygo/gosmsn/client/session"
)

func main() {
	// 启动记录数据库的程序
	if sms.Conf.GetString("Mongo.URI") != "" {
		sms.PersistenceSmsJournal()
	}

	phone := flag.String("p", "13800001111", "phone")
	message := flag.String("m", "hello world", "message")
	iterates := flag.Int("i", 1000, "iterates")
	flag.Parse()

	i := *iterates
	for i > 0 {
		sms.Send(*message, *phone)
		i--
	}

	destroy()

	// 连续10次缓存大小为0，则退出程序
	var x, k = 0, 10
	for x > 0 || k > 0 {
		x = 0
		session.SequenceIdResultCacheMap.Range(func(key, value any) bool {
			x++
			return true
		})
		if x == 0 {
			k--
		}
		log.Infof("Current Cache size is %d, and k is %d", x, k)
		time.Sleep(time.Second)
	}

	log.Warn("main goroutine exit.")
	log.Sync()
	os.Exit(0)
}

func destroy() {
	go func() {
		// 接收服务停止信号
		<-bs.StatChan()
		log.Sync()
		os.Exit(0)
	}()
}
