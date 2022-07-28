package server

import (
	"fmt"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"

	"github.com/hrygo/gosmsn/my_log"
)

var logDb = my_log.New("CounterDB")

func (s *Server) OnClose(c gnet.Conn, e error) (action gnet.Action) {
	msg := fmt.Sprintf("[%s] OnClose ===", s.name)
	sc := Session(c)
	defer deferFunc(c, s, sc)

	fields := FlatMapLog(sc.LogSession(), s.LogCounter(), []log.Field{OpConnectionClose.Field(), ErrorField(e)})
	log.Warn(msg, fields...)
	logDb.Info("save counter", FlatMapLog(sc.LogSession(), sc.LogCounter())...)
	_ = logDb.Sync() // 强制刷盘
	return
}

func deferFunc(c gnet.Conn, s *Server, sc *session) {
	sc.closePoolChan() // 关闭通道，释放线程池
	s.sessionPool.Delete(sc.Id())
	s.activeSessions -= 1
	c.SetContext(nil)
	if sc != nil {
		sc.conn = nil
		sc = nil
	}
}
