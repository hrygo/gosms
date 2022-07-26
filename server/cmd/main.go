package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/hrygo/log"

	bs "github.com/hrygo/gosmsn/bootstrap"
	"github.com/hrygo/gosmsn/server"
)

func main() {
	rand.Seed(time.Now().Unix()) // 随机种子
	log.Infof("current pid is %s.", savePid("_.pid"))
	pprofDebug()

	server.Start(server.New(server.CMPP))

	// 接收服务停止信号
	<-bs.StatChan()
	log.Warn("main goroutine exit.")
	log.Sync()
	os.Exit(0)
}

// 在程序执行的当前目录生成pid文件
func savePid(pf string) string {
	file, err := os.OpenFile(pf, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Errorf("%v", err)
	}
	pid := fmt.Sprintf("%d", os.Getpid())
	writer := bufio.NewWriter(file)

	defer func(file *os.File, writer *bufio.Writer) {
		_ = writer.Flush()
		_ = file.Close()
	}(file, writer)

	_, _ = writer.WriteString(pid)

	return pid
}

// 开启pprof，监听请求
func pprofDebug() {
	if bs.ConfigYml.GetBool("AppDebug") {
		go func() {
			var pprof = bs.ConfigYml.GetInt("Server.Pprof.Port")
			log.Warnf("debug pprof on http://localhost:%d/debug/pprof/", pprof)
			if err := http.ListenAndServe(fmt.Sprintf(":%d", pprof), nil); err != nil {
				log.Fatalf("start pprof failed on %s", pprof)
			}
		}()
	}
}
