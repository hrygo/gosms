package sms

import (
	"context"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosmsn/client/session"
	"github.com/hrygo/gosmsn/codec"
	db "github.com/hrygo/gosmsn/databse"
)

// Send sms to phones return query id
func Send(message string, phones ...string) (queryId int64) {
	if len(phones) < 1 {
		return
	}
	queryId = codec.B64Seq.NextVal()
	var probLen = len(phones) * (len(message)/70 + 1)
	var results = make([]any, 0, probLen)
	for _, phone := range phones {
		phone := phone
		sc := SelectSession(phone)
		for sc == nil {
			time.Sleep(time.Millisecond)
			sc = SelectSession(phone)
		}
		results = append(results, sc.Send(phone, message)...)
		sc.AddCounter()
	}
	saveQueryCache(queryId, results)
	return
}

func Query(queryId int64) []any {
	value, ok := resultQueryCacheMap.Load(queryId)
	if ok {
		result, ok := value.([]any)
		if ok {
			_, ok := result[0].(*session.Result)
			if ok {
				return result
			}
		}
	}
	return nil
}

func saveQueryCache(key int64, value []any) {
	resultQueryCacheMap.Store(key, value)
}

func PersistenceSmsJournal() {
	db.InitDB(Conf, "Mongo")
	coll := db.Collection("smsdb", "journal")
	StartCacheExpireTicker(func(results []any) {
		_ = AsyncPool().Submit(func() {
			log.Infof("[Persistence] Save %d send results to db.", len(results))
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			many, err := coll.InsertMany(ctx, results)
			if err != nil {
				return
			}
			log.Infof("[Persistence] Save to mongodb success: %v", many)
		})
	})
}
