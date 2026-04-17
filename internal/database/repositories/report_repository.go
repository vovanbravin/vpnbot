package repositories

import (
	"context"
	"errors"
	"tgbot/internal/database"
	"tgbot/internal/database/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ReportRepository struct {
	collection *mongo.Collection
}

func NewReportRepository(db *database.MongoDB) *ReportRepository {
	return &ReportRepository{db.GetCollection("reports")}
}

func (r *ReportRepository) Insert(ctx context.Context, report *models.Report) error {
	_, err := r.collection.InsertOne(ctx, report)
	return err
}

func (r *ReportRepository) GetAllReportByUserId(ctx context.Context, userId int64) ([]models.Report, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userId})

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var reports []models.Report
	if err = cursor.All(ctx, &reports); err != nil {
		return nil, err
	}

	return reports, nil
}

func (r *ReportRepository) GetReportByNumber(ctx context.Context, number int64) (*models.Report, error) {

	var report models.Report
	filter := bson.M{"number": number}
	err := r.collection.FindOne(ctx, filter).Decode(&report)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &report, nil
}

func (r *ReportRepository) CountByStatus(ctx context.Context, status models.ReportStatus) (int, error) {
	filter := bson.M{"status": status}
	cursor, err := r.collection.Find(ctx, filter)

	defer cursor.Close(ctx)

	if err != nil {
		return 0, err
	}

	var results []models.Report

	if err = cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	return len(results), nil
}

func (r *ReportRepository) GetAllByStatus(ctx context.Context, status models.ReportStatus) ([]models.Report, error) {
	return r.GetAllByStatuses(ctx, []models.ReportStatus{status})
}

func (r *ReportRepository) GetAllByStatuses(ctx context.Context, statuses []models.ReportStatus) ([]models.Report, error) {
	filter := bson.M{"status": bson.M{"$in": statuses}}
	cursor, err := r.collection.Find(ctx, filter)

	defer cursor.Close(ctx)

	if err != nil {
		return nil, err
	}

	var reports []models.Report

	if err = cursor.All(ctx, &reports); err != nil {
		return nil, err
	}

	return reports, nil
}

func (r *ReportRepository) GetAllByStatusesAndUserId(ctx context.Context, userId int64, statuses []models.ReportStatus) ([]models.Report, error) {
	filter := bson.M{"status": bson.M{"$in": statuses}, "user_id": userId}
	cursor, err := r.collection.Find(ctx, filter)

	defer cursor.Close(ctx)

	if err != nil {
		return nil, err
	}

	var reports []models.Report

	if err = cursor.All(ctx, &reports); err != nil {
		return nil, err
	}

	return reports, nil
}

func (r *ReportRepository) Update(ctx context.Context, report models.Report) error {
	filter := bson.M{"_id": report.ID}
	update := bson.M{"$set": report}
	_, err := r.collection.UpdateOne(ctx, filter, update)

	return err
}
