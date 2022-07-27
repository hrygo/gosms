package server

import (
	"fmt"

	"github.com/hrygo/log"
	"github.com/panjf2000/gnet/v2"
)

func (s *Server) OnClose(c gnet.Conn, e error) (action gnet.Action) {
	msg := fmt.Sprintf("[%s] OnClose ===", s.name)
	ses := Session(c)
	defer deferFunc(c, s, ses)

	fields := FlatMapLog(ses.LogSession(), s.LogCounter(), []log.Field{OpConnectionClose.Field(), ErrorField(e)})
	log.Warn(msg, fields...)
	return
}

func deferFunc(c gnet.Conn, s *Server, ses *session) {
	ses.closePoolChan() // 关闭通道，释放线程池
	s.sessions.Delete(ses.Id())
	c.SetContext(nil)
	if ses != nil {
		ses.conn = nil
		ses = nil
	}
}
