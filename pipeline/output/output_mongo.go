package output

import (
	"github.com/LSvKing/centipede/config"
	"github.com/LSvKing/centipede/items"
	"upper.io/db.v3"
	"upper.io/db.v3/mongo"
	//"fmt"
)

type (
	OutPutMongGo struct {
	}

	Collections map[string]db.Collection
)

func (this *OutPutMongGo) OutPut(dataCache items.DataCache) {
	log.Debug("mongo")

	appConfig := config.Get()

	var settings = mongo.ConnectionURL{
		Host:     appConfig.Mongo.Host,     // server IP.
		Database: appConfig.Mongo.Database, // Database name.
	}

	sess, err := mongo.Open(settings)

	if err != nil {
		log.Fatalf("db.Open(): %q\n", err)
	}

	defer sess.Close() // Remember to close the database session.

	collections := make(Collections)

	for _, value := range dataCache {

		if _, ok := collections[value.Collection]; !ok {
			collections[value.Collection] = sess.Collection(value.Collection)
		}

		data := make(map[string]interface{})

		for _, v := range value.Data {
			data[v.Field] = v.Value
		}

		collections[value.Collection].Insert(data)
	}
}
