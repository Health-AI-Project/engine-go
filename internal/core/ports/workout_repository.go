package ports

import (
	"context"
	"time"

	"healthai/engine/internal/core/domain"
)

type WorkoutRepository interface {
	// Exercise Catalog
	GetExercises(ctx context.Context, limit int) ([]domain.Exercise, error)
	CreateExercise(ctx context.Context, exercise *domain.Exercise) error
	GetExercisesByEquipAndDiff(ctx context.Context, equipment []domain.EquipmentType, difficulty []domain.DifficultyLevel) ([]domain.Exercise, error)

	// User Workout Usage (for progression logic)
	GetLastWorkouts(ctx context.Context, userID string, limit int) ([]domain.Workout, error)

	// Workout Plans
	CreateWorkoutPlan(ctx context.Context, plan *domain.WorkoutPlan) error
	GetWorkoutPlanByDate(ctx context.Context, userID string, date time.Time) (*domain.WorkoutPlan, error)
}
