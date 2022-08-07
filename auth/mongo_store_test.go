package auth

import (
	"context"
	"testing"

	"github.com/hrygo/log"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	db "github.com/hrygo/gosmsn/databse"
)

func TestMongoStore_FindByCid(t *testing.T) {
	mcache := &MongoStore{cache: make(map[string]*Client)}
	mcache.Load()

	c := mcache.FindByCid("cmpp", "123456")
	assert.True(t, c != nil && c.ClientId == "123456")
}

// 将文件中的配置信息迁移到MongoDB中
func TestMongoStore_Load(t *testing.T) {
	TestDeleteAll(t)
	cache := &YamlStore{cache: make(map[string]*Client)}
	cache.Load()

	for _, client := range cache.cache {
		saveToMongo(client)
	}

	mcache := &MongoStore{cache: make(map[string]*Client)}
	mcache.Load()

	assert.True(t, len(cache.cache) == len(mcache.cache))
}

func TestDeleteAll(t *testing.T) {
	coll := db.Mongo.Client.Database(dbname).Collection(collection)
	rod := db.Mongo.Config.GetDuration(db.Mongo.Prefix + ".ReadTimeout")
	ctx, cancel := context.WithTimeout(context.Background(), rod)
	defer cancel()
	many, err := coll.DeleteMany(ctx, bson.D{})
	if err != nil {
		return
	}
	log.Warnf("deleted: %v", many)
}

func saveToMongo(client *Client) {
	coll := db.Mongo.Client.Database(dbname).Collection(collection)
	rod := db.Mongo.Config.GetDuration(db.Mongo.Prefix + ".ReadTimeout")
	ctx, cancel := context.WithTimeout(context.Background(), rod)
	defer cancel()

	one, err := coll.InsertOne(ctx, client)
	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Infof("saved: %v", one)
}
