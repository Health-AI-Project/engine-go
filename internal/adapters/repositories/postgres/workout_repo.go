package postgres

import (
	"context"
	"time"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/ports"

	"gorm.io/gorm"
)

type WorkoutRepository struct {
	db *gorm.DB
}

func NewWorkoutRepository(db *gorm.DB) ports.WorkoutRepository {
	return &WorkoutRepository{db: db}
}

func (r *WorkoutRepository) GetExercises(ctx context.Context, limit int) ([]domain.Exercise, error) {
	var exercises []domain.Exercise
	err := r.db.WithContext(ctx).Limit(limit).Find(&exercises).Error
	return exercises, err
}

func (r *WorkoutRepository) CreateExercise(ctx context.Context, exercise *domain.Exercise) error {
	return r.db.WithContext(ctx).Create(exercise).Error
}

func (r *WorkoutRepository) GetExercisesByEquipAndDiff(ctx context.Context, equipment []domain.EquipmentType, difficulty []domain.DifficultyLevel) ([]domain.Exercise, error) {
	var exercises []domain.Exercise
	query := r.db.WithContext(ctx)

	if len(equipment) > 0 {
		query = query.Where("required_equipment IN ?", equipment)
	}
	if len(difficulty) > 0 {
		query = query.Where("difficulty IN ?", difficulty)
	}

	err := query.Find(&exercises).Error
	return exercises, err
}

func (r *WorkoutRepository) GetLastWorkouts(ctx context.Context, userID string, limit int) ([]domain.Workout, error) {
	var workouts []domain.Workout
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("date desc").
		Limit(limit).
		Find(&workouts).Error
	return workouts, err
}

func (r *WorkoutRepository) CreateWorkoutPlan(ctx context.Context, plan *domain.WorkoutPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *WorkoutRepository) GetWorkoutPlanByDate(ctx context.Context, userID string, date time.Time) (*domain.WorkoutPlan, error) {
	var plan domain.WorkoutPlan
	err := r.db.WithContext(ctx).
		Preload("Exercises").
		Preload("Exercises.Exercise").
		Where("user_id = ? AND date = ?", userID, date).
		First(&plan).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}
