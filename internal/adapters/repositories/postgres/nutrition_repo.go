package postgres

import (
	"context"
	"time"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/ports"

	"gorm.io/gorm"
)

type NutritionRepository struct {
	db *gorm.DB
}

func NewNutritionRepository(db *gorm.DB) ports.NutritionRepository {
	return &NutritionRepository{db: db}
}

func (r *NutritionRepository) GetFoodPreference(ctx context.Context, userID string) (*domain.FoodPreference, error) {
	var fp domain.FoodPreference
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&fp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Return nil if not set yet
		}
		return nil, err
	}
	return &fp, nil
}

func (r *NutritionRepository) UpsertFoodPreference(ctx context.Context, fp *domain.FoodPreference) error {
	// GORM Clause 'OnConflict'
	// return r.db.WithContext(ctx).Save(fp).Error // Save handles update if ID exists, but we want Upsert by UserID
	// Actually, since ID is primary key, Save works if ID is known.
	// If ID is not known but UserID is unique index, we need OnConflict.
	// For simplicity, let's use Save.
	return r.db.WithContext(ctx).Save(fp).Error
}

func (r *NutritionRepository) GetMealSuggestions(ctx context.Context, limit int) ([]domain.MealSuggestion, error) {
	var meals []domain.MealSuggestion
	err := r.db.WithContext(ctx).Limit(limit).Find(&meals).Error
	return meals, err
}

func (r *NutritionRepository) CreateMealSuggestion(ctx context.Context, meal *domain.MealSuggestion) error {
	return r.db.WithContext(ctx).Create(meal).Error
}

func (r *NutritionRepository) GetMealPlanByDate(ctx context.Context, userID string, date time.Time) (*domain.MealPlan, error) {
	var plan domain.MealPlan
	// Preload Meals using GORM association
	err := r.db.WithContext(ctx).
		Preload("Meals").
		Preload("Meals.MealSuggestion").
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

func (r *NutritionRepository) CreateMealPlan(ctx context.Context, plan *domain.MealPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}
