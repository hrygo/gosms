package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hrygo/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/hrygo/gosms/msc_server"
)

func TestNew(t *testing.T) {
	cLog := msc.New("test",
		log.WithCaller(true),
		log.AddStacktrace(log.ErrorLevel),
		zap.Hooks(printCallerHook),
	)

	cLog.Info("cLog test", log.String("hello", "world"))
	cLog.Error("cLog test", log.String("hello", "world"))

	time.Sleep(time.Millisecond)
}

func printCallerHook(ze zapcore.Entry) error {
	go func(entry zapcore.Entry) {
		fmt.Printf("PrintCallerHook: %s\n", entry.Caller.TrimmedPath())
	}(ze)
	return nil
}
