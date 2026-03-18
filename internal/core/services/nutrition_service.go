package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/ports"
)

type NutritionService struct {
	repo     ports.NutritionRepository
	userRepo ports.UserRepository
}

func NewNutritionService(repo ports.NutritionRepository, userRepo ports.UserRepository) *NutritionService {
	return &NutritionService{repo: repo, userRepo: userRepo}
}

func (s *NutritionService) GetFoodPreferences(ctx context.Context, userID string) (*domain.FoodPreference, error) {
	return s.repo.GetFoodPreference(ctx, userID)
}

func (s *NutritionService) UpdateFoodPreferences(ctx context.Context, userID string, allergies []string, diet domain.DietType, disliked []string) error {
	prefs, err := s.repo.GetFoodPreference(ctx, userID)
	if err != nil {
		// If error is not found, we create new
		prefs = &domain.FoodPreference{
			UserID: userID,
		}
	} else if prefs == nil {
		prefs = &domain.FoodPreference{
			UserID: userID,
		}
	}

	prefs.Allergies = allergies
	prefs.DietType = diet
	prefs.DislikedIngredients = disliked
	prefs.UpdatedAt = time.Now()

	return s.repo.UpsertFoodPreference(ctx, prefs)
}

type MealAnalysisReport struct {
	IsBalanced     bool
	Warnings       []string // e.g., "Too many carbs", "Low protein"
	CriticalAlerts []string // e.g., "Contains Peanut (Allergen)"
}

func (s *NutritionService) AnalyzeMeal(ctx context.Context, userID string, mealInput domain.MealSuggestion) (*MealAnalysisReport, error) {
	report := &MealAnalysisReport{
		IsBalanced:     true,
		Warnings:       []string{},
		CriticalAlerts: []string{},
	}

	// 1. Fetch User Preferences (for Allergies)
	prefs, err := s.repo.GetFoodPreference(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user preferences: %w", err)
	}

	// 2. Check Allergies
	if prefs != nil {
		for _, ingredient := range mealInput.Ingredients {
			for _, allergy := range prefs.Allergies {
				if ingredient == allergy {
					report.CriticalAlerts = append(report.CriticalAlerts, fmt.Sprintf("Contains %s (Allergen)", allergy))
					report.IsBalanced = false
				}
			}
		}
	}

	// 3. Check Macros (Simplified logic)
	// Example rule: > 40% of standard 2000kcal day in one meal is "Heavy" (800kcal)
	if mealInput.Calories > 800 {
		report.Warnings = append(report.Warnings, "High calorie meal (>800kcal)")
	}

	// Check Protein (Arbitrary threshold for example: < 10g is low)
	if mealInput.Protein < 10 {
		report.Warnings = append(report.Warnings, "Low protein content (<10g)")
	}

	if len(report.Warnings) > 0 {
		report.IsBalanced = false // debatable, but let's flag it
	}

	return report, nil
}

func (s *NutritionService) GenerateDailyPlan(ctx context.Context, userID string, targetDate time.Time) (*domain.MealPlan, error) {
	// 1. Fetch Constraints
	prefs, err := s.repo.GetFoodPreference(ctx, userID)
	if err != nil {
		return nil, err
	}

	dietType := domain.DietTypeNone
	if prefs != nil {
		dietType = prefs.DietType
	}

	// 2. Determine Targets (Simplified: could come from HealthProfile BMR calc)
	// For now, hardcode or assume defaults
	targetCalories := 2200.0
	targetProtein := 150.0

	// 3. Fetch Meal Catalog
	suggestions, err := s.repo.GetMealSuggestions(ctx, 100) // fetch catalogue
	if err != nil {
		return nil, err
	}

	// 4. Algorithm: Select Meals
	// Filter by DietType & Allergies
	var validMeals []domain.MealSuggestion
	for _, meal := range suggestions {
		// Filter diet type
		if dietType != domain.DietTypeNone {
			matchesDiet := false
			for _, tag := range meal.DietTags {
				if tag == string(dietType) {
					matchesDiet = true
					break
				}
			}
			if !matchesDiet {
				continue
			}
		}

		// Filter Allergies
		isAllergic := false
		if prefs != nil {
			for _, ing := range meal.Ingredients {
				for _, all := range prefs.Allergies {
					if ing == all {
						isAllergic = true
						break
					}
				}
			}
		}
		if !isAllergic {
			validMeals = append(validMeals, meal)
		}
	}

	if len(validMeals) < 3 {
		return nil, fmt.Errorf("not enough valid meals found for diet: %s", dietType)
	}

	// Simple heuristic: Pick 3 random meals (Breakfast, Lunch, Dinner)
	// In production -> Knapsack problem to fit calories
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	plan := &domain.MealPlan{
		ID:             "", // GORM hooks will handle UUID
		UserID:         userID,
		Date:           targetDate,
		TargetCalories: targetCalories,
		TargetProtein:  targetProtein,
		Meals:          []domain.MealPlanItem{},
	}

	// Pick 3 unique meals if possible
	perm := r.Perm(len(validMeals))

	mealTypes := []domain.MealType{domain.MealTypeBreakfast, domain.MealTypeLunch, domain.MealTypeDinner}

	for i, idx := range perm {
		if i >= 3 {
			break
		}
		meal := validMeals[idx]
		plan.Meals = append(plan.Meals, domain.MealPlanItem{
			// ID handled by DB/GORM usually, but we are creating the object structure
			MealSuggestionID: meal.ID, // DB Relationship
			MealSuggestion:   meal,    // Struct embedding for return
			MealType:         mealTypes[i],
		})
	}

	// 5. Save Plan
	err = s.repo.CreateMealPlan(ctx, plan)
	if err != nil {
		return nil, err
	}

	return plan, nil
}
