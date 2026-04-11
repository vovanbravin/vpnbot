package repositories

import (
	"context"
	"sync"
	"tgbot/internal/database"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CounterRepository struct {
	collection *mongo.Collection
	mu         sync.Mutex
}

func NewCounterRepository(db *database.MongoDB) *CounterRepository {
	return &CounterRepository{collection: db.GetCollection("counter")}
}

func (r *CounterRepository) GetNextNumber(ctx context.Context, name string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"value": 1}}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var counter struct {
		Value int64
	}

	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)

	if err != nil {
		return 0, err
	}

	return counter.Value, nil
}
