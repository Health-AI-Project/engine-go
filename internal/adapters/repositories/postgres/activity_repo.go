package postgres

import (
	"context"
	"errors"
	"time"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/ports"

	"gorm.io/gorm"
)

type ActivityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) ports.ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) GetDailyLogByDate(ctx context.Context, userID string, date time.Time) (*domain.DailyLog, error) {
	var log domain.DailyLog
	// Search by truncated date to ensure day-level uniqueness
	// Using GORM, depending on driver, we might need strict date comparison.
	// Assuming incoming date is already truncated to midnight by service, but good to be safe.
	// Here simple equality check on date column.
	if err := r.db.WithContext(ctx).Where("user_id = ? AND date = ?", userID, date).First(&log).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if not found, let service handle logic
		}
		return nil, err
	}
	return &log, nil
}

func (r *ActivityRepository) CreateDailyLog(ctx context.Context, log *domain.DailyLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *ActivityRepository) UpdateDailyLog(ctx context.Context, log *domain.DailyLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

func (r *ActivityRepository) GetDailyLogHistory(ctx context.Context, userID string, days int) ([]domain.DailyLog, error) {
	var logs []domain.DailyLog
	since := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -days)
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND date >= ?", userID, since).
		Order("date ASC").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *ActivityRepository) CreateWorkout(ctx context.Context, workout *domain.Workout) error {
	return r.db.WithContext(ctx).Create(workout).Error
}
