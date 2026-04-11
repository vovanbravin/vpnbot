package repositories

import (
	"context"
	"errors"
	"tgbot/internal/database"
	"tgbot/internal/database/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *database.MongoDB) *UserRepository {
	return &UserRepository{collection: db.GetCollection("users")}
}

func (r *UserRepository) CreateUser(ctx context.Context, user models.User) error {
	_, err := r.collection.InsertOne(ctx, user)

	return err
}

func (r *UserRepository) GetUserByTelegramId(ctx context.Context, telegram_id int64) (*models.User, error) {
	var user models.User
	filter := bson.M{"telegram_id": telegram_id}
	err := r.collection.FindOne(ctx, filter).Decode(&user)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}
