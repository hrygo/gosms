package session

import (
	"time"

	"github.com/hrygo/gosms/codec"
)

func (s *Session) sendBySgip(phone string, message string, options ...codec.OptionFunc) []any {
	return nil
}

func (s *Session) onTrafficSgip(cmd, seq uint32, buff []byte) {
	s.activeTime = time.Now()
}
