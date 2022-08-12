package sgip

import (
	"fmt"
	"sync"
	"time"
)

var Sequencer *SequenceNumber

type SequenceNumber struct {
	sync.Mutex
	Node      uint32 // 节点编号
	timestamp uint32 // 时间戳 mmddhhmmss 十进制数
	sequence  uint32 // 循序号
}

func (s *SequenceNumber) NextVal() (rs []uint32) {
	s.Lock()
	defer s.Unlock()
	s.sequence = (s.sequence + 1) & 0xffffffff
	now := time.Now()

	rs = make([]uint32, 3)
	rs[0] = s.Node
	s.timestamp = uint32(now.Month() * 100000000)
	s.timestamp += uint32(now.Day() * 1000000)
	s.timestamp += uint32(now.Hour() * 10000)
	s.timestamp += uint32(now.Minute() * 100)
	s.timestamp += uint32(now.Second())
	rs[1] = s.timestamp
	rs[2] = s.sequence
	return
}

func (s *SequenceNumber) CurVal() (rs []uint32) {
	rs = make([]uint32, 3)
	rs[0] = s.Node
	rs[1] = s.timestamp
	rs[2] = s.sequence
	return rs
}

func (s *SequenceNumber) String() string {
	return fmt.Sprintf("%010d%010d%08x", s.Node, s.timestamp, s.sequence)
}
