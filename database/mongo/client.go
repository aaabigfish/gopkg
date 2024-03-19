package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	MC *mongo.Client
)

const (
	defaultTimeout = 50 * time.Second
	maxPoolSize    = 10
)

type Client struct {
	C BaseCollection
}

func MongoClient(dbName, colName string, mc *mongo.Client) *Client {
	dataBase := mc.Database(dbName)
	c := &BaseCollectionImpl{
		DbName:     dbName,
		ColName:    colName,
		DataBase:   dataBase,
		Collection: dataBase.Collection(colName),
	}
	client := &Client{}
	client.C = c
	return client
}

func InitMongo(mongoUrl string) (*mongo.Client, error) {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	MC, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoUrl).SetMaxPoolSize(maxPoolSize))
	if err != nil {
		return nil, err
	}
	if err := MC.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	return MC, nil
}
