package services

import (
	"context"
	"time"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/ports"
)

type ActivityService struct {
	repo ports.ActivityRepository
}

func NewActivityService(repo ports.ActivityRepository) *ActivityService {
	return &ActivityService{repo: repo}
}

// LogNutrition adds nutrition data to today's log.
// Logic: Find existing today ? update : create new.
func (s *ActivityService) LogNutrition(ctx context.Context, userID string, calories, protein, carbs, fat float64) error {
	today := time.Now().Truncate(24 * time.Hour) // Normalize to midnight

	log, err := s.repo.GetDailyLogByDate(ctx, userID, today)
	if err != nil {
		return err
	}

	if log == nil {
		// Create new
		newLog := domain.NewDailyLog(userID, today)
		newLog.TotalProtein = protein
		newLog.TotalCarbs = carbs
		newLog.TotalFat = fat
		newLog.TotalCalories = s.calculateCalories(calories, protein, carbs, fat)
		return s.repo.CreateDailyLog(ctx, newLog)
	}

	// Update existing
	log.TotalProtein += protein
	log.TotalCarbs += carbs
	log.TotalFat += fat
	log.TotalCalories += s.calculateCalories(calories, protein, carbs, fat) 
	log.UpdatedAt = time.Now()

	return s.repo.UpdateDailyLog(ctx, log)
}

func (s *ActivityService) LogWorkout(ctx context.Context, userID string, wType string, duration int, calories float64) error {
	workout := domain.NewWorkout(userID, domain.WorkoutType(wType), duration, calories)
	return s.repo.CreateWorkout(ctx, workout)
}

// calculateCalories helper: prefers specific calorie input, else estimates from macros
// However, task says "Recalcule automatiquement le total calorique après chaque ajout."
// This implies aggregation.
// For the *delta* being added: 
func (s *ActivityService) calculateCalories(explicitCal, p, c, f float64) float64 {
	if explicitCal > 0 {
		return explicitCal
	}
	// Fallback estimation: 4 kcal/g protein/carbs, 9 kcal/g fat
	return (p * 4) + (c * 4) + (f * 9)
}
