package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewClient(ctx context.Context, connUri string) (*mongo.Client, error) {
	return mongo.Connect(ctx, options.Client().ApplyURI(connUri))
}
