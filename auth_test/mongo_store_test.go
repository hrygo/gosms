package auth

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hrygo/log"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	db "github.com/hrygo/gosms/database/mongodb"

	"github.com/hrygo/gosms/auth"
)

var pass bool

func init() {
	db.InitDB(ConfigYml, "Mongo")
	pass = db.Mongo == nil
}

// 将文件中的配置信息迁移到MongoDB中
func TestMongoStore_Load(t *testing.T) {
	if pass {
		return
	}

	deleteAll()
	cache := &auth.YamlStore{
		Cache:  make(map[string]*auth.Client),
		Config: ConfigYml,
	}
	cache.Load()

	for _, client := range cache.Cache {
		saveToMongo(client)
	}

	mcache := &auth.MongoStore{
		Cache:  make(map[string]*auth.Client),
		Config: ConfigYml,
	}
	mcache.Load()

	assert.True(t, len(cache.Cache) == len(mcache.Cache))
}

func TestMongoStore_FindByCid(t *testing.T) {
	if pass {
		return
	}
	mcache := &auth.MongoStore{
		Cache:  make(map[string]*auth.Client),
		Config: ConfigYml,
	}
	mcache.Load()

	c := mcache.FindByCid("cmpp", "123456")
	assert.True(t, c != nil && c.ClientId == "123456")

	for _, c := range mcache.Cache {
		jsn, err := json.MarshalIndent(c, "", "    ")
		assert.True(t, err == nil)
		t.Log("\n" + string(jsn))
	}
}

func deleteAll() {
	if pass {
		return
	}

	coll := db.Mongo.Client.Database(auth.DBN).Collection(auth.Collection)
	rod := db.Mongo.Config.GetDuration(db.Mongo.Prefix + ".ReadTimeout")
	ctx, cancel := context.WithTimeout(context.Background(), rod)
	defer cancel()
	many, err := coll.DeleteMany(ctx, bson.D{})
	if err != nil {
		return
	}
	log.Warnf("deleted: %v", many)
}

func saveToMongo(client *auth.Client) {
	coll := db.Mongo.Client.Database(auth.DBN).Collection(auth.Collection)
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
