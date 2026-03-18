package services_test

import (
	"context"
	"testing"
	"time"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/ports/mocks"
	"healthai/engine/internal/core/services"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetFoodPreferences(t *testing.T) {
	mockNutritionRepo := new(mocks.MockNutritionRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := services.NewNutritionService(mockNutritionRepo, mockUserRepo)

	t.Run("Returns existing preferences", func(t *testing.T) {
		expected := &domain.FoodPreference{
			UserID:   "user-1",
			DietType: domain.DietTypeVegan,
			Allergies: pq.StringArray{"Peanut"},
		}
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-1").Return(expected, nil).Once()

		result, err := service.GetFoodPreferences(context.Background(), "user-1")
		assert.NoError(t, err)
		assert.Equal(t, domain.DietTypeVegan, result.DietType)
		mockNutritionRepo.AssertExpectations(t)
	})

	t.Run("Returns nil when no preferences", func(t *testing.T) {
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-none").Return(nil, nil).Once()

		result, err := service.GetFoodPreferences(context.Background(), "user-none")
		assert.NoError(t, err)
		assert.Nil(t, result)
		mockNutritionRepo.AssertExpectations(t)
	})
}

func TestUpdateFoodPreferences(t *testing.T) {
	mockNutritionRepo := new(mocks.MockNutritionRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := services.NewNutritionService(mockNutritionRepo, mockUserRepo)

	t.Run("Create new preferences when none exist", func(t *testing.T) {
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-new").Return(nil, nil).Once()
		mockNutritionRepo.On("UpsertFoodPreference", context.Background(), mock.MatchedBy(func(fp *domain.FoodPreference) bool {
			return fp.UserID == "user-new" && fp.DietType == domain.DietTypeKeto && len(fp.Allergies) == 1 && len(fp.DislikedIngredients) == 1
		})).Return(nil).Once()

		err := service.UpdateFoodPreferences(context.Background(), "user-new", []string{"Gluten"}, domain.DietTypeKeto, []string{"Onion"})
		assert.NoError(t, err)
		mockNutritionRepo.AssertExpectations(t)
	})

	t.Run("Update existing preferences", func(t *testing.T) {
		existing := &domain.FoodPreference{
			UserID:   "user-exist",
			DietType: domain.DietTypeNone,
		}
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-exist").Return(existing, nil).Once()
		mockNutritionRepo.On("UpsertFoodPreference", context.Background(), mock.MatchedBy(func(fp *domain.FoodPreference) bool {
			return fp.DietType == domain.DietTypeVegetarian
		})).Return(nil).Once()

		err := service.UpdateFoodPreferences(context.Background(), "user-exist", []string{}, domain.DietTypeVegetarian, []string{})
		assert.NoError(t, err)
		mockNutritionRepo.AssertExpectations(t)
	})
}

func TestAnalyzeMeal(t *testing.T) {
	mockNutritionRepo := new(mocks.MockNutritionRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := services.NewNutritionService(mockNutritionRepo, mockUserRepo)

	t.Run("Balanced meal with no allergens", func(t *testing.T) {
		prefs := &domain.FoodPreference{
			Allergies: pq.StringArray{},
		}
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-1").Return(prefs, nil).Once()

		meal := domain.MealSuggestion{
			Calories:    500,
			Protein:     30,
			Ingredients: pq.StringArray{"Chicken", "Rice"},
		}
		report, err := service.AnalyzeMeal(context.Background(), "user-1", meal)
		assert.NoError(t, err)
		assert.True(t, report.IsBalanced)
		assert.Empty(t, report.CriticalAlerts)
		mockNutritionRepo.AssertExpectations(t)
	})

	t.Run("High calorie meal -> Warning", func(t *testing.T) {
		prefs := &domain.FoodPreference{
			Allergies: pq.StringArray{},
		}
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-1").Return(prefs, nil).Once()

		meal := domain.MealSuggestion{
			Calories:    900,
			Protein:     50,
			Ingredients: pq.StringArray{"Pasta", "Cheese"},
		}
		report, err := service.AnalyzeMeal(context.Background(), "user-1", meal)
		assert.NoError(t, err)
		assert.False(t, report.IsBalanced)
		assert.Contains(t, report.Warnings, "High calorie meal (>800kcal)")
		mockNutritionRepo.AssertExpectations(t)
	})

	t.Run("Low protein meal -> Warning", func(t *testing.T) {
		prefs := &domain.FoodPreference{
			Allergies: pq.StringArray{},
		}
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-1").Return(prefs, nil).Once()

		meal := domain.MealSuggestion{
			Calories:    400,
			Protein:     5,
			Ingredients: pq.StringArray{"Salad"},
		}
		report, err := service.AnalyzeMeal(context.Background(), "user-1", meal)
		assert.NoError(t, err)
		assert.Contains(t, report.Warnings, "Low protein content (<10g)")
		mockNutritionRepo.AssertExpectations(t)
	})

	t.Run("Allergen detected -> Critical alert", func(t *testing.T) {
		prefs := &domain.FoodPreference{
			Allergies: pq.StringArray{"Peanut"},
		}
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-allergy").Return(prefs, nil).Once()

		meal := domain.MealSuggestion{
			Calories:    500,
			Protein:     25,
			Ingredients: pq.StringArray{"Peanut", "Rice"},
		}
		report, err := service.AnalyzeMeal(context.Background(), "user-allergy", meal)
		assert.NoError(t, err)
		assert.False(t, report.IsBalanced)
		assert.Len(t, report.CriticalAlerts, 1)
		assert.Contains(t, report.CriticalAlerts[0], "Peanut")
		mockNutritionRepo.AssertExpectations(t)
	})

	t.Run("Nil preferences -> no allergen check crash", func(t *testing.T) {
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-nil").Return(nil, nil).Once()

		meal := domain.MealSuggestion{
			Calories:    500,
			Protein:     25,
			Ingredients: pq.StringArray{"Chicken"},
		}
		report, err := service.AnalyzeMeal(context.Background(), "user-nil", meal)
		assert.NoError(t, err)
		assert.True(t, report.IsBalanced)
		mockNutritionRepo.AssertExpectations(t)
	})

	t.Run("Repo error -> propagates", func(t *testing.T) {
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-err").Return(nil, assert.AnError).Once()

		meal := domain.MealSuggestion{}
		_, err := service.AnalyzeMeal(context.Background(), "user-err", meal)
		assert.Error(t, err)
		mockNutritionRepo.AssertExpectations(t)
	})
}

func TestGenerateDailyPlan(t *testing.T) {
	mockNutritionRepo := new(mocks.MockNutritionRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := services.NewNutritionService(mockNutritionRepo, mockUserRepo)

	targetDate := time.Now().AddDate(0, 0, 1)

	t.Run("Not enough valid meals -> error", func(t *testing.T) {
		prefs := &domain.FoodPreference{
			DietType:  domain.DietTypeVegan,
			Allergies: pq.StringArray{},
		}
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-1").Return(prefs, nil).Once()

		meals := []domain.MealSuggestion{
			{ID: "m1", DietTags: pq.StringArray{"KETO"}, Ingredients: pq.StringArray{"Steak"}},
		}
		mockNutritionRepo.On("GetMealSuggestions", context.Background(), 100).Return(meals, nil).Once()

		_, err := service.GenerateDailyPlan(context.Background(), "user-1", targetDate)
		assert.Error(t, err)
		mockNutritionRepo.AssertExpectations(t)
	})

	t.Run("No preferences -> uses DietTypeNone", func(t *testing.T) {
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-nopref").Return(nil, nil).Once()

		meals := []domain.MealSuggestion{
			{ID: "m1", Ingredients: pq.StringArray{"A"}, DietTags: pq.StringArray{}},
			{ID: "m2", Ingredients: pq.StringArray{"B"}, DietTags: pq.StringArray{}},
			{ID: "m3", Ingredients: pq.StringArray{"C"}, DietTags: pq.StringArray{}},
			{ID: "m4", Ingredients: pq.StringArray{"D"}, DietTags: pq.StringArray{}},
		}
		mockNutritionRepo.On("GetMealSuggestions", context.Background(), 100).Return(meals, nil).Once()
		mockNutritionRepo.On("CreateMealPlan", context.Background(), mock.AnythingOfType("*domain.MealPlan")).Return(nil).Once()

		result, err := service.GenerateDailyPlan(context.Background(), "user-nopref", targetDate)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Meals, 3)
		mockNutritionRepo.AssertExpectations(t)
	})

	t.Run("GetFoodPreference error -> propagates", func(t *testing.T) {
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-err").Return(nil, assert.AnError).Once()

		_, err := service.GenerateDailyPlan(context.Background(), "user-err", targetDate)
		assert.Error(t, err)
		mockNutritionRepo.AssertExpectations(t)
	})

	t.Run("GetMealSuggestions error -> propagates", func(t *testing.T) {
		mockNutritionRepo.On("GetFoodPreference", context.Background(), "user-2").Return(nil, nil).Once()
		mockNutritionRepo.On("GetMealSuggestions", context.Background(), 100).Return(nil, assert.AnError).Once()

		_, err := service.GenerateDailyPlan(context.Background(), "user-2", targetDate)
		assert.Error(t, err)
		mockNutritionRepo.AssertExpectations(t)
	})
}
