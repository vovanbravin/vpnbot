package models

import "go.mongodb.org/mongo-driver/v2/bson"

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID         bson.ObjectID `bson:"_id"`
	TelegramId int64         `bson:"telegram_id"`
	Role       UserRole      `bson:"role"`
}
