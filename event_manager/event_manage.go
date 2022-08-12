package event_manager

import (
	"strings"
	"sync"

	"github.com/hrygo/log"
)

const (
	ErrorsFuncEventAlreadyExists = "注册函数类事件失败，键名已经被注册"
	ErrorsFuncEventNotRegister   = "没有找到键名对应的函数"
	ShutdownHookerPrefix         = "graceful_shutdown_hooker_"
)

// Event 事件为无返回值的任意函数
type Event = func(args ...any)

// 定义一个全局事件存储变量，本模块只负责存储 键 => 函数
var sMap sync.Map

// 定义一个事件管理结构体
type eventManage struct {
	prefix string
}

// CreateEventManage 创建一个事件管理工厂
func CreateEventManage(prefix string) *eventManage {
	return &eventManage{prefix: prefix}
}

// RegisterShutdownHooker 注册优雅停机钩子事件
func RegisterShutdownHooker(key string, hookers ...Event) bool {
	return CreateEventManage(ShutdownHookerPrefix).Register(key, hookers...)
}

// RegisterShutdownHookerAddChan 注册优雅停机钩子事件，并额外返回一个取消信号
// 当执行优雅停机时，events 会先被执行，然后发送信号给通道
func RegisterShutdownHookerAddChan(key string, events ...Event) <-chan bool {
	cancel := make(chan bool, 1)
	events = append(events, func(args ...any) {
		cancel <- true
	})
	ok := RegisterShutdownHooker(key, events...)
	if !ok {
		close(cancel)
		return nil
	}
	return cancel
}

// GracefullyShutdown 执行优雅停机钩子事件
func GracefullyShutdown() {
	CreateEventManage(ShutdownHookerPrefix).FuzzyCall()
}

// Register 注册事件
func (m *eventManage) Register(key string, events ...Event) bool {
	if len(events) < 1 {
		return false
	}
	if !strings.HasPrefix(key, m.prefix) {
		key = m.prefix + key
	}
	// 判断key下是否已有事件
	if _, exists := m.Get(key); exists == false {
		sMap.Store(key, func(args ...any) {
			for _, fn := range events {
				fn()
			}
		})
		return true
	} else {
		log.Info(ErrorsFuncEventAlreadyExists + " , 相关键名：" + key)
	}
	return false
}

// Get 获取事件
func (m *eventManage) Get(key string) (Event, bool) {
	if !strings.HasPrefix(key, m.prefix) {
		key = m.prefix + key
	}
	if value, exists := sMap.Load(key); exists {
		e, ok := value.(Event)
		return e, ok
	}
	return nil, false
}

// Call 执行事件
func (m *eventManage) Call(key string, args ...interface{}) {
	if !strings.HasPrefix(key, m.prefix) {
		key = m.prefix + key
	}
	if fn, exists := m.Get(key); exists {
		log.Infof("Call: %#v, Key: %s", fn, key)
		fn(args)
		log.Infof("Done: %#v, Key: %s", fn, key)

	} else {
		log.Error(ErrorsFuncEventNotRegister + ", 键名：" + key)
	}
}

// Delete 删除事件
func (m *eventManage) Delete(key string) {
	if !strings.HasPrefix(key, m.prefix) {
		key = m.prefix + key
	}
	sMap.Delete(key)
}

// FuzzyCall 根据键的前缀，模糊调用。 仅适用于无参数事件，请谨慎使用.
func (m *eventManage) FuzzyCall() {
	sMap.Range(func(key, value interface{}) bool {
		if keyName, ok := key.(string); ok {
			if strings.HasPrefix(keyName, m.prefix) {
				m.Call(keyName)
			}
		}
		return true
	})
}
