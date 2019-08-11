package users

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrUserNotFound = errors.New("user not found")

type Store interface {
	GetUserByTelegramUserID(ctx context.Context, tgUserID int64) (User, error)
	StoreUser(ctx context.Context, user *User) error
}

type Mongo struct {
	usersColl *mongo.Collection
}

func NewMongo(usersColl *mongo.Collection) Mongo {
	return Mongo{usersColl: usersColl}
}

func (m Mongo) GetUserByTelegramUserID(ctx context.Context, tgUserID int64) (User, error) {
	u := User{}
	err := m.usersColl.FindOne(ctx, bson.M{"TelegramUserID": tgUserID}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		err = ErrUserNotFound
	}
	return u, err
}

func (m Mongo) StoreUser(ctx context.Context, user *User) error {
	_, err := m.usersColl.InsertOne(ctx, user)
	return err
}
