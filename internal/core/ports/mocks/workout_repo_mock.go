package mocks

import (
	"context"
	"time"

	"healthai/engine/internal/core/domain"

	"github.com/stretchr/testify/mock"
)

type MockWorkoutRepository struct {
	mock.Mock
}

func (m *MockWorkoutRepository) GetExercises(ctx context.Context, limit int) ([]domain.Exercise, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Exercise), args.Error(1)
}

func (m *MockWorkoutRepository) CreateExercise(ctx context.Context, exercise *domain.Exercise) error {
	args := m.Called(ctx, exercise)
	return args.Error(0)
}

func (m *MockWorkoutRepository) GetExercisesByEquipAndDiff(ctx context.Context, equipment []domain.EquipmentType, difficulty []domain.DifficultyLevel) ([]domain.Exercise, error) {
	args := m.Called(ctx, equipment, difficulty)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Exercise), args.Error(1)
}

func (m *MockWorkoutRepository) GetLastWorkouts(ctx context.Context, userID string, limit int) ([]domain.Workout, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Workout), args.Error(1)
}

func (m *MockWorkoutRepository) CreateWorkoutPlan(ctx context.Context, plan *domain.WorkoutPlan) error {
	args := m.Called(ctx, plan)
	return args.Error(0)
}

func (m *MockWorkoutRepository) GetWorkoutPlanByDate(ctx context.Context, userID string, date time.Time) (*domain.WorkoutPlan, error) {
	args := m.Called(ctx, userID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WorkoutPlan), args.Error(1)
}
