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
			return log.TotalCalories == 700
		})).Return(nil).Once()

		err := service.LogNutrition(context.Background(), userID, 500, 0, 0, 0)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Calorie fallback from macros when explicit is 0", func(t *testing.T) {
		mockRepo.On("GetDailyLogByDate", context.Background(), userID, mock.AnythingOfType("time.Time")).Return(nil, nil).Once()

		mockRepo.On("CreateDailyLog", context.Background(), mock.MatchedBy(func(log *domain.DailyLog) bool {
			expected := (20.0 * 4) + (50.0 * 4) + (10.0 * 9)
			return log.UserID == userID && log.TotalCalories == expected
		})).Return(nil).Once()

		err := service.LogNutrition(context.Background(), userID, 0, 20, 50, 10)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Aggregation with macro fallback on existing log", func(t *testing.T) {
		existingLog := &domain.DailyLog{
			UserID:        userID,
			Date:          today,
			TotalCalories: 100,
			TotalProtein:  5,
			TotalCarbs:    10,
			TotalFat:      2,
		}

		mockRepo.On("GetDailyLogByDate", context.Background(), userID, mock.AnythingOfType("time.Time")).Return(existingLog, nil).Once()

		mockRepo.On("UpdateDailyLog", context.Background(), mock.MatchedBy(func(log *domain.DailyLog) bool {
			delta := (10.0 * 4) + (25.0 * 4) + (5.0 * 9)
			return log.TotalCalories == 100+delta &&
				log.TotalProtein == 5+10 &&
				log.TotalCarbs == 10+25 &&
				log.TotalFat == 2+5
		})).Return(nil).Once()

		err := service.LogNutrition(context.Background(), userID, 0, 10, 25, 5)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetDailyLogByDate error -> Should propagate", func(t *testing.T) {
		mockRepo.On("GetDailyLogByDate", context.Background(), "err-user", mock.AnythingOfType("time.Time")).Return(nil, assert.AnError).Once()

		err := service.LogNutrition(context.Background(), "err-user", 100, 0, 0, 0)
		assert.Error(t, err)
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

	t.Run("Strength workout", func(t *testing.T) {
		mockRepo.On("CreateWorkout", context.Background(), mock.MatchedBy(func(w *domain.Workout) bool {
			return w.UserID == userID && w.Type == domain.WorkoutTypeStrength && w.Duration == 45 && w.CaloriesBurned == 400
		})).Return(nil).Once()

		err := service.LogWorkout(context.Background(), userID, "STRENGTH", 45, 400)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetDailyStats(t *testing.T) {
	mockRepo := new(mocks.MockActivityRepository)
	service := services.NewActivityService(mockRepo)

	userID := "test-user-id"

	t.Run("Returns stats for today", func(t *testing.T) {
		expectedLog := &domain.DailyLog{
			UserID:        userID,
			TotalCalories: 1500,
			TotalProtein:  100,
			TotalCarbs:    200,
			TotalFat:      50,
		}
		mockRepo.On("GetDailyLogByDate", context.Background(), userID, mock.AnythingOfType("time.Time")).Return(expectedLog, nil).Once()

		result, err := service.GetDailyStats(context.Background(), userID)
		assert.NoError(t, err)
		assert.Equal(t, 1500.0, result.TotalCalories)
		assert.Equal(t, 100.0, result.TotalProtein)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Returns nil when no log exists", func(t *testing.T) {
		mockRepo.On("GetDailyLogByDate", context.Background(), "new-user", mock.AnythingOfType("time.Time")).Return(nil, nil).Once()

		result, err := service.GetDailyStats(context.Background(), "new-user")
		assert.NoError(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

