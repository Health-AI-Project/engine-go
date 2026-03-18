package services_test

import (
	"context"
	"testing"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/ports/mocks"
	"healthai/engine/internal/core/services"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateWorkout(t *testing.T) {
	mockWorkoutRepo := new(mocks.MockWorkoutRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	service := services.NewWorkoutService(mockWorkoutRepo, mockUserRepo)

	userID := "test-user"

	t.Run("Generates a workout plan with valid exercises", func(t *testing.T) {
		mockWorkoutRepo.On("GetLastWorkouts", context.Background(), userID, 5).Return(nil, nil).Once()

		exercises := []domain.Exercise{
			{ID: "ex1", Name: "Push-up", Type: domain.ExerciseTypeStrength, Contraindications: pq.StringArray{}},
			{ID: "ex2", Name: "Squat", Type: domain.ExerciseTypeStrength, Contraindications: pq.StringArray{}},
			{ID: "ex3", Name: "Plank", Type: domain.ExerciseTypeStrength, Contraindications: pq.StringArray{}},
		}
		mockWorkoutRepo.On("GetExercisesByEquipAndDiff", context.Background(),
			[]domain.EquipmentType{domain.EquipmentNone},
			[]domain.DifficultyLevel{domain.DifficultyBeginner},
		).Return(exercises, nil).Once()

		mockWorkoutRepo.On("CreateWorkoutPlan", context.Background(), mock.AnythingOfType("*domain.WorkoutPlan")).Return(nil).Once()

		constraints := services.WorkoutConstraints{
			DurationMinutes: 30,
			Equipment:       []domain.EquipmentType{domain.EquipmentNone},
			UserInjuries:    []string{},
		}

		plan, err := service.GenerateWorkout(context.Background(), userID, constraints)
		assert.NoError(t, err)
		assert.NotNil(t, plan)
		assert.Equal(t, 30, plan.DurationMinutes)
		assert.Equal(t, "PENDING", plan.Status)
		assert.True(t, len(plan.Exercises) > 0)
		mockWorkoutRepo.AssertExpectations(t)
	})

	t.Run("Filters out exercises with contraindications matching injuries", func(t *testing.T) {
		mockWorkoutRepo.On("GetLastWorkouts", context.Background(), userID, 5).Return(nil, nil).Once()

		exercises := []domain.Exercise{
			{ID: "ex1", Name: "Push-up", Contraindications: pq.StringArray{"ShoulderInjury"}},
			{ID: "ex2", Name: "Squat", Contraindications: pq.StringArray{}},
			{ID: "ex3", Name: "Plank", Contraindications: pq.StringArray{}},
		}
		mockWorkoutRepo.On("GetExercisesByEquipAndDiff", context.Background(),
			[]domain.EquipmentType{domain.EquipmentNone},
			[]domain.DifficultyLevel{domain.DifficultyBeginner},
		).Return(exercises, nil).Once()

		mockWorkoutRepo.On("CreateWorkoutPlan", context.Background(), mock.AnythingOfType("*domain.WorkoutPlan")).Return(nil).Once()

		constraints := services.WorkoutConstraints{
			DurationMinutes: 20,
			Equipment:       []domain.EquipmentType{domain.EquipmentNone},
			UserInjuries:    []string{"ShoulderInjury"},
		}

		plan, err := service.GenerateWorkout(context.Background(), userID, constraints)
		assert.NoError(t, err)
		for _, item := range plan.Exercises {
			assert.NotEqual(t, "ex1", item.ExerciseID)
		}
		mockWorkoutRepo.AssertExpectations(t)
	})

	t.Run("No exercises found after filtering -> error", func(t *testing.T) {
		mockWorkoutRepo.On("GetLastWorkouts", context.Background(), userID, 5).Return(nil, nil).Once()

		exercises := []domain.Exercise{
			{ID: "ex1", Name: "Push-up", Contraindications: pq.StringArray{"ShoulderInjury"}},
		}
		mockWorkoutRepo.On("GetExercisesByEquipAndDiff", context.Background(),
			[]domain.EquipmentType{domain.EquipmentNone},
			[]domain.DifficultyLevel{domain.DifficultyBeginner},
		).Return(exercises, nil).Once()

		constraints := services.WorkoutConstraints{
			DurationMinutes: 20,
			Equipment:       []domain.EquipmentType{domain.EquipmentNone},
			UserInjuries:    []string{"ShoulderInjury"},
		}

		_, err := service.GenerateWorkout(context.Background(), userID, constraints)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no exercises found")
		mockWorkoutRepo.AssertExpectations(t)
	})

	t.Run("Repo GetExercises error -> propagates", func(t *testing.T) {
		mockWorkoutRepo.On("GetLastWorkouts", context.Background(), userID, 5).Return(nil, nil).Once()
		mockWorkoutRepo.On("GetExercisesByEquipAndDiff", context.Background(),
			[]domain.EquipmentType{domain.EquipmentDumbbells},
			[]domain.DifficultyLevel{domain.DifficultyBeginner},
		).Return(nil, assert.AnError).Once()

		constraints := services.WorkoutConstraints{
			DurationMinutes: 30,
			Equipment:       []domain.EquipmentType{domain.EquipmentDumbbells},
			UserInjuries:    []string{},
		}

		_, err := service.GenerateWorkout(context.Background(), userID, constraints)
		assert.Error(t, err)
		mockWorkoutRepo.AssertExpectations(t)
	})

	t.Run("Short duration -> at least 1 exercise", func(t *testing.T) {
		mockWorkoutRepo.On("GetLastWorkouts", context.Background(), userID, 5).Return(nil, nil).Once()

		exercises := []domain.Exercise{
			{ID: "ex1", Name: "Push-up", Contraindications: pq.StringArray{}},
		}
		mockWorkoutRepo.On("GetExercisesByEquipAndDiff", context.Background(),
			[]domain.EquipmentType{domain.EquipmentNone},
			[]domain.DifficultyLevel{domain.DifficultyBeginner},
		).Return(exercises, nil).Once()

		mockWorkoutRepo.On("CreateWorkoutPlan", context.Background(), mock.AnythingOfType("*domain.WorkoutPlan")).Return(nil).Once()

		constraints := services.WorkoutConstraints{
			DurationMinutes: 5,
			Equipment:       []domain.EquipmentType{domain.EquipmentNone},
			UserInjuries:    []string{},
		}

		plan, err := service.GenerateWorkout(context.Background(), userID, constraints)
		assert.NoError(t, err)
		assert.Len(t, plan.Exercises, 1)
		mockWorkoutRepo.AssertExpectations(t)
	})

	t.Run("With progression check (5+ recent workouts)", func(t *testing.T) {
		recentWorkouts := make([]domain.Workout, 5)
		mockWorkoutRepo.On("GetLastWorkouts", context.Background(), userID, 5).Return(recentWorkouts, nil).Once()

		exercises := []domain.Exercise{
			{ID: "ex1", Name: "Push-up", Contraindications: pq.StringArray{}},
			{ID: "ex2", Name: "Squat", Contraindications: pq.StringArray{}},
		}
		mockWorkoutRepo.On("GetExercisesByEquipAndDiff", context.Background(),
			[]domain.EquipmentType{domain.EquipmentNone},
			[]domain.DifficultyLevel{domain.DifficultyBeginner},
		).Return(exercises, nil).Once()

		mockWorkoutRepo.On("CreateWorkoutPlan", context.Background(), mock.AnythingOfType("*domain.WorkoutPlan")).Return(nil).Once()

		constraints := services.WorkoutConstraints{
			DurationMinutes: 25,
			Equipment:       []domain.EquipmentType{domain.EquipmentNone},
			UserInjuries:    []string{},
		}

		plan, err := service.GenerateWorkout(context.Background(), userID, constraints)
		assert.NoError(t, err)
		assert.NotNil(t, plan)
		mockWorkoutRepo.AssertExpectations(t)
	})
}
