package snowflake

import (
	"sync"
	"time"

	"github.com/panjf2000/gnet/v2/pkg/logging"
)

const (
	epoch           = int64(1577808000000)                       // 设置起始时间(时间戳/毫秒)：2020-01-01 00:00:00，有效期69年
	timestampBits   = uint(41)                                   // 时间戳占用位数
	datacenterBits  = uint(2)                                    // 数据中心id所占位数
	workerBits      = uint(7)                                    // 机器id所占位数
	sequenceBits    = uint(12)                                   // 序列所占的位数
	timestampMax    = int64(-1 ^ (-1 << timestampBits))          // 时间戳最大值
	sequenceMask    = int64(-1 ^ (-1 << sequenceBits))           // 支持的最大序列id数量
	workerShift     = sequenceBits                               // 机器id左移位数
	datacenterShift = sequenceBits + workerBits                  // 数据中心id左移位数
	timestampShift  = sequenceBits + workerBits + datacenterBits // 时间戳左移位数
)

var log = logging.GetDefaultLogger()

type Snowflake struct {
	sync.Mutex         // 锁
	timestamp    int64 // 时间戳 ，毫秒
	workerId     int64 // 工作节点
	datacenterId int64 // 数据中心机房id
	sequence     int64 // 序列号
}

func NewSnowflake(d int64, w int64) *Snowflake {
	return &Snowflake{datacenterId: d, workerId: w}
}
func (s *Snowflake) NextVal() int64 {
	s.Lock()
	defer s.Unlock()
	now := time.Now().UnixNano() / 1000000 // 转毫秒
	if s.timestamp == now {
		// 当同一时间戳（精度：毫秒）下多次生成id会增加序列号
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// 如果当前序列超出12bit长度，则需要等待下一毫秒
			// 下一毫秒将使用sequence:0
			for now <= s.timestamp {
				time.Sleep(time.Nanosecond)
				now = time.Now().UnixNano() / 1000000
			}
		}
	} else {
		// 不同时间戳（精度：毫秒）下直接使用序列号：0
		s.sequence = 0
	}
	t := now - epoch
	if t > timestampMax {
		log.Errorf("epoch must be between 0 and %d", timestampMax-1)
		return 0
	}
	s.timestamp = now
	r := (t << timestampShift) | (s.datacenterId << datacenterShift) | (s.workerId << workerShift) | (s.sequence)
	return r
}
