package entity_storage_mongodb

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/gwlog"
)

const (
	DEFAULT_DB_NAME = "goworld"
)

var (
	db *mgo.Database
)

type MongoDBEntityStorge struct {
	db *mgo.Database
}

func OpenMongoDB(url string, dbname string) (*MongoDBEntityStorge, error) {
	gwlog.Debug("Connecting MongoDB ...")
	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)
	if dbname == "" {
		// if db is not specified, use default
		dbname = DEFAULT_DB_NAME
	}
	db = session.DB(dbname)
	return &MongoDBEntityStorge{
		db: db,
	}, nil
}

func collectionName(name string) string {
	return fmt.Sprintf("%s", name)
}

func (ss *MongoDBEntityStorge) Write(typeName string, entityID common.EntityID, data interface{}) error {
	col := ss.getCollection(typeName)
	_, err := col.UpsertId(entityID, bson.M{
		"data": data,
	})
	return err
}

func (ss *MongoDBEntityStorge) Read(typeName string, entityID common.EntityID) (interface{}, error) {
	col := ss.getCollection(typeName)
	q := col.FindId(entityID)
	var doc bson.M
	err := q.One(&doc)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}(doc["data"].(bson.M)), nil
}

func (ss *MongoDBEntityStorge) getCollection(typeName string) *mgo.Collection {
	return ss.db.C(typeName)
}

func (ss *MongoDBEntityStorge) List(typeName string) ([]common.EntityID, error) {
	col := ss.getCollection(typeName)
	var docs []bson.M
	err := col.Find(nil).Select(bson.M{"_id": 1}).All(&docs)
	if err != nil {
		return nil, err
	}

	entityIDs := make([]common.EntityID, len(docs))
	for i, doc := range docs {
		entityIDs[i] = common.EntityID(doc["_id"].(string))
	}
	return entityIDs, nil
}

func (ss *MongoDBEntityStorge) Exists(typeName string, entityID common.EntityID) (bool, error) {
	col := ss.getCollection(typeName)
	query := col.FindId(entityID)
	var doc bson.M
	err := query.One(&doc)
	if err == nil {
		// doc found
		return true, nil
	} else if err == mgo.ErrNotFound {
		return false, nil
	} else {
		return false, err
	}
}
