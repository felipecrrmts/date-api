package persistence

import "go.mongodb.org/mongo-driver/mongo"

type Database struct {
	*User
}

func New(db *mongo.Database) *Database {
	return &Database{
		User: NewItemPersistence(db),
	}
}
