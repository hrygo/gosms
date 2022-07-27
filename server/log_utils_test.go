package server_test

import (
	"testing"

	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/server"
)

func TestFlatMapLog(t *testing.T) {
	s1 := []log.Field{
		log.Int("1", 1),
		log.Int("1", 2),
	}

	var s2 []log.Field

	var s3 = log.Int("3", 1)

	log.Info("flatmap: ", server.FlatMapLog(s1, s2, []log.Field{s3})...)
}
