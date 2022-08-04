package session

import (
	"time"
)

func (s *Session) onTrafficSgip(cmd, seq uint32, buff []byte) {
	s.activeTime = time.Now()
}
