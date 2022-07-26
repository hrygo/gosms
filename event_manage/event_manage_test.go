package event_manage

import (
	"log"
	"testing"
)

func TestEventManage(t *testing.T) {
	em := CreateEventManage("test_event_")

	em.Register("println", log.Println)
	em.Call("println", "hello", "world")

	em.Register("no_args_f1", f1)
	em.Register("no_args_f2", f2)
	em.Register("no_args_f3", f3)
	em.Register("no_args_f3", f3)
	em.FuzzyCall()

	em.Delete("no_args_f3")
	em.Call("no_args_f3")

	em.FuzzyCall()

	fn, exits := em.Get("println")
	if exits {
		fn("hello world end.")
	}
}

var f1 = func(args ...interface{}) {
	log.Println("call f1")
}

var f2 = func(args ...interface{}) {
	log.Println("call f2")
}

var f3 = func(args ...interface{}) {
	log.Println("call f3")
}
