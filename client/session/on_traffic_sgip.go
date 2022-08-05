package session

import (
	"time"
)

func (s *Session) sendBySgip(phone string, message string) []*Result {
	return nil
}

func (s *Session) onTrafficSgip(cmd, seq uint32, buff []byte) {
	s.activeTime = time.Now()
}
