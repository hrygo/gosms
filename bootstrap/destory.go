package bs

import (
  "fmt"
  "os"
  "os/signal"
  "syscall"

  "github.com/hrygo/log"

  "github.com/hrygo/gosmsn/event_manage"
)

// 优雅停机相关代码

const ShutdownEventPrefix = "graceful_shutdown_"

func init() {
  go func() {
    // 控制关闭 main goroutine
    defer func() { statChan <- struct{}{} }()
    // 接收信号
    var c = make(chan os.Signal)
    defer close(c)

    // 监听可能的退出信号
    signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
    // 接收信号管道中的值
    received := <-c

    log.Warn(fmt.Sprintf("收到信号 [%s] 进程即将结束！", received.String()))

    // 优雅停机的善后处理
    event_manage.CreateEventManageFactory().FuzzyCall(ShutdownEventPrefix)
  }()
}
