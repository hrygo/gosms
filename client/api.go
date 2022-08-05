package sms

import (
	"github.com/hrygo/gosmsn/client/session"
	"github.com/hrygo/gosmsn/codec"
)

// Send sms to phones return query id
func Send(message string, phones ...string) (queryId int64) {
	if len(phones) < 1 {
		return
	}
	queryId = codec.B64Seq.NextVal()
	var probLen = len(phones) * (len(message)/70 + 1)
	var results = make([]*session.Result, 0, probLen)
	for _, phone := range phones {
		phone := phone
		sc := SelectSession(phone)
		if sc != nil {
			results = append(results, sc.Send(phone, message)...)
		}
	}
	saveQueryCache(queryId, results)
	return
}

func Query(queryId int64) []*session.Result {
	value, ok := resultQueryCacheMap.Load(queryId)
	if ok {
		result, ok := value.([]*session.Result)
		if ok {
			return result
		}
	}
	return nil
}

func saveQueryCache(key int64, value []*session.Result) {
	resultQueryCacheMap.Store(key, value)
}
