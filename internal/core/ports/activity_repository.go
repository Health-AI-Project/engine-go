package ports

import (
	"context"
	"time"

	"healthai/engine/internal/core/domain"
)

type ActivityRepository interface {
	// Daily Log
	GetDailyLogByDate(ctx context.Context, userID string, date time.Time) (*domain.DailyLog, error)
	CreateDailyLog(ctx context.Context, log *domain.DailyLog) error
	UpdateDailyLog(ctx context.Context, log *domain.DailyLog) error

	// Workout
	CreateWorkout(ctx context.Context, workout *domain.Workout) error
}
