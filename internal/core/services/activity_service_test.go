package services_test

import (
	"context"
	"testing"
	"time"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/ports/mocks"
	"healthai/engine/internal/core/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogNutrition(t *testing.T) {
	mockRepo := new(mocks.MockActivityRepository)
	service := services.NewActivityService(mockRepo)

	userID := "test-user-id"
	today := time.Now().Truncate(24 * time.Hour)

	t.Run("New Entry -> Should create new log", func(t *testing.T) {
		mockRepo.On("GetDailyLogByDate", context.Background(), userID, mock.AnythingOfType("time.Time")).Return(nil, nil).Once()
		
		mockRepo.On("CreateDailyLog", context.Background(), mock.MatchedBy(func(log *domain.DailyLog) bool {
			return log.UserID == userID && log.TotalCalories == 500
		})).Return(nil).Once()

		err := service.LogNutrition(context.Background(), userID, 500, 30, 50, 20)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Existing Entry -> Should aggregate", func(t *testing.T) {
		existingLog := &domain.DailyLog{
			UserID:        userID,
			Date:          today,
			TotalCalories: 200,
			TotalProtein:  10,
			TotalCarbs:    20,
			TotalFat:      5,
		}

		mockRepo.On("GetDailyLogByDate", context.Background(), userID, mock.AnythingOfType("time.Time")).Return(existingLog, nil).Once()

		mockRepo.On("UpdateDailyLog", context.Background(), mock.MatchedBy(func(log *domain.DailyLog) bool {
			// Initial 200 + Added 500 = 700
			return log.TotalCalories == 700
		})).Return(nil).Once()

		err := service.LogNutrition(context.Background(), userID, 500, 0, 0, 0)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestLogWorkout(t *testing.T) {
	mockRepo := new(mocks.MockActivityRepository)
	service := services.NewActivityService(mockRepo)

	userID := "test-user-id"

	t.Run("Should create workout log", func(t *testing.T) {
		mockRepo.On("CreateWorkout", context.Background(), mock.MatchedBy(func(w *domain.Workout) bool {
			return w.UserID == userID && w.Type == domain.WorkoutTypeCardio && w.Duration == 30
		})).Return(nil).Once()

		err := service.LogWorkout(context.Background(), userID, "CARDIO", 30, 300)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
