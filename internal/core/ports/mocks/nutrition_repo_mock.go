package mocks

import (
	"context"
	"time"

	"healthai/engine/internal/core/domain"

	"github.com/stretchr/testify/mock"
)

type MockNutritionRepository struct {
	mock.Mock
}

func (m *MockNutritionRepository) GetFoodPreference(ctx context.Context, userID string) (*domain.FoodPreference, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FoodPreference), args.Error(1)
}

func (m *MockNutritionRepository) UpsertFoodPreference(ctx context.Context, fp *domain.FoodPreference) error {
	args := m.Called(ctx, fp)
	return args.Error(0)
}

func (m *MockNutritionRepository) GetMealSuggestions(ctx context.Context, limit int) ([]domain.MealSuggestion, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MealSuggestion), args.Error(1)
}

func (m *MockNutritionRepository) CreateMealSuggestion(ctx context.Context, meal *domain.MealSuggestion) error {
	args := m.Called(ctx, meal)
	return args.Error(0)
}

func (m *MockNutritionRepository) GetMealPlanByDate(ctx context.Context, userID string, date time.Time) (*domain.MealPlan, error) {
	args := m.Called(ctx, userID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MealPlan), args.Error(1)
}

func (m *MockNutritionRepository) CreateMealPlan(ctx context.Context, plan *domain.MealPlan) error {
	args := m.Called(ctx, plan)
	return args.Error(0)
}
