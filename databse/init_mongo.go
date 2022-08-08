package db

import (
	"context"
	"os"
	"strings"

	"github.com/hrygo/log"
	"github.com/hrygo/yaml_config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type database struct {
	Client *mongo.Client
	Config yaml_config.YmlConfig
	Prefix string
}

var Mongo *database

func InitDB(config yaml_config.YmlConfig, prefix string) {
	if "" == config.GetString(prefix+".URI") {
		return
	}
	Mongo = &database{
		Client: setConnect(config, prefix),
		Config: config,
		Prefix: prefix,
	}
}

func Collection(db, coll string) *mongo.Collection {
	return Mongo.Client.Database(db).Collection(coll)
}

func setConnect(config yaml_config.YmlConfig, prefix string) *mongo.Client {
	user := os.Getenv("MONGO_USER")
	passwd := os.Getenv("MONGO_PASSWD")
	if user == "" || passwd == "" {
		log.Panic("Environment variable MONGO_USER or MONGO_PASSWD not set!")
	}

	uri := config.GetString(prefix + ".URI")
	uri = strings.ReplaceAll(uri, "<user>", user)
	uri = strings.ReplaceAll(uri, "<passwd>", passwd)

	connectTimeout := config.GetDuration(prefix + ".ConnectTimeout")
	heartbeatInterval := config.GetDuration(prefix + ".HeartbeatInterval")
	minPoolSize := config.GetInt(prefix + ".MinPoolSize")
	maxPoolSize := config.GetInt(prefix + ".MaxPoolSize")

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	option := options.Client()
	option.ApplyURI(uri)
	option.SetMinPoolSize(uint64(minPoolSize))
	option.SetMaxPoolSize(uint64(maxPoolSize))
	option.SetHeartbeatInterval(heartbeatInterval)
	// TODO 可增加更多设置

	client, err := mongo.Connect(ctx, option)
	if err != nil {
		log.Panic(err.Error())
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Panic(err.Error())
	}
	log.Infof("Connected to %s", uri)
	return client
}
