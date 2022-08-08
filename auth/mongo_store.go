package auth

import (
	"context"
	"strings"

	"github.com/hrygo/log"
	"go.mongodb.org/mongo-driver/bson"

	db "github.com/hrygo/gosms/databse"
)

const (
	dbname     = "smsdb"
	collection = "authenticatedClients"
)

type MongoStore storage

func (m *MongoStore) Load() {
	coll := db.Mongo.Client.Database(dbname).Collection(collection)
	rod := db.Mongo.Config.GetDuration(db.Mongo.Prefix + ".ReadTimeout")
	ctx, cancel := context.WithTimeout(context.Background(), rod)
	defer cancel()

	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		log.Error(err.Error())
		return
	}

	m.Lock()
	defer m.Unlock()
	for cursor.Next(ctx) {
		c := &Client{}
		err := cursor.Decode(c)
		if err != nil {
			return
		}
		m.cache[c.ISP+"_"+c.ClientId] = c
	}
}

func (m *MongoStore) FindByCid(isp string, cid string) (c *Client) {
	m.Lock()
	defer m.Unlock()
	client := m.cache[strings.ToLower(isp)+"_"+cid]
	return client
}
