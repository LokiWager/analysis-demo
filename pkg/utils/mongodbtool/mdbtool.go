package mongodbtool

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	DefaultDBName = "diagnostic"
)

var (
	client *mongo.Client
)

func getClient() *mongo.Client {
	if client != nil {
		return client
	}
	var err error
	client, err = mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	return client
}

func GetCollection(collectionName string) *mongo.Collection {
	return getClient().Database(DefaultDBName).Collection(collectionName)
}

func CloseMDB() {
	if client != nil {
		_ = client.Disconnect(nil)
	}
}
