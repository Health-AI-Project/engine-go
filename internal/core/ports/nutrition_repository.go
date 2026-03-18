package ports

import (
	"context"
	"time"

	"healthai/engine/internal/core/domain"
)

type NutritionRepository interface {
	// Food Preferences
	GetFoodPreference(ctx context.Context, userID string) (*domain.FoodPreference, error)
	UpsertFoodPreference(ctx context.Context, fp *domain.FoodPreference) error

	// Meal Suggestions (Catalog)
	GetMealSuggestions(ctx context.Context, limit int) ([]domain.MealSuggestion, error)
	CreateMealSuggestion(ctx context.Context, meal *domain.MealSuggestion) error

	// Meal Plans
	GetMealPlanByDate(ctx context.Context, userID string, date time.Time) (*domain.MealPlan, error)
	CreateMealPlan(ctx context.Context, plan *domain.MealPlan) error
}
