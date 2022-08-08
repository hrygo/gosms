package event_manage

import (
	"strings"
	"sync"

	"github.com/hrygo/log"

	"github.com/hrygo/gosms/my_errors"
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

// Register 注册事件
func (m *eventManage) Register(key string, e Event) bool {
	if !strings.HasPrefix(key, m.prefix) {
		key = m.prefix + key
	}
	// 判断key下是否已有事件
	if _, exists := m.Get(key); exists == false {
		sMap.Store(key, e)
		return true
	} else {
		log.Info(my_errors.ErrorsFuncEventAlreadyExists + " , 相关键名：" + key)
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
		fn(args)
	} else {
		log.Error(my_errors.ErrorsFuncEventNotRegister + ", 键名：" + key)
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
