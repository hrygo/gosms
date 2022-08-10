package sms

import (
	"context"
	"time"

	"github.com/hrygo/log"

	"github.com/hrygo/gosms/database/mongodb"

	"github.com/hrygo/gosms/smc_client/session"

	"github.com/hrygo/gosms/codec"
)

func Send(message, phone string, options ...codec.OptionFunc) (queryId int64) {
	return SendN(message, []string{phone}, options...)
}

// SendN sms to phones return query id
func SendN(message string, phones []string, options ...codec.OptionFunc) (queryId int64) {
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
		results = append(results, sc.Send(phone, message, options...)...)
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
	for _, a := range value {
		r := a.(*session.Result)
		r.QueryId = key
	}
	resultQueryCacheMap.Store(key, value)
}

func PersistenceSmsJournal() {
	mongodb.InitDB(ConfigYml, "Mongo")
	coll := mongodb.Collection("smsdb", "journal")
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
