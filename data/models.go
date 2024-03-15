package data

import (
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client

func New(mongo *mongo.Client) Models {
	client = mongo

	return Models {
		RefreshToken: RefreshToken{},
	}
}

type Models struct {
	RefreshToken RefreshToken
}

type RefreshToken struct {
	Token string `bson:refresh_token json:"refresh_token"`
	UserID string `bson:_id json:"user_id"`
}
