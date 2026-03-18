package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/ports"
)

type WorkoutService struct {
	repo     ports.WorkoutRepository
	userRepo ports.UserRepository
}

func NewWorkoutService(repo ports.WorkoutRepository, userRepo ports.UserRepository) *WorkoutService {
	return &WorkoutService{repo: repo, userRepo: userRepo}
}

type WorkoutConstraints struct {
	DurationMinutes int
	Equipment       []domain.EquipmentType
	UserInjuries    []string
	TargetMuscle    []string // optional
}

func (s *WorkoutService) GenerateWorkout(ctx context.Context, userID string, constraints WorkoutConstraints) (*domain.WorkoutPlan, error) {
	// 1. Get User Profile for Fitness Level (mocked as Intermediate if null)
	// In real app, fetch from User Domain or HealthProfile
	// For now, let's assume "BEGINNER" as default
	userLevel := domain.DifficultyBeginner

	// 2. Progression Check: Check last 5 workouts
	lastWorkouts, err := s.repo.GetLastWorkouts(ctx, userID, 5)
	if err == nil && len(lastWorkouts) >= 5 {
		// Logique naive: Si tous les derniers workouts sont marqués "Easy" (imaginaire), on up.
		// Since we don't have "Rating" in domain yet, let's skip strict progression logic implementation
		// and just adhere to the userLevel.
		// TODO: Implement rating check.
	}

	// 3. Fetch Candidate Exercises
	// Filter by Equipment (Repo Level) and Roughly by Difficulty
	candidates, err := s.repo.GetExercisesByEquipAndDiff(ctx, constraints.Equipment, []domain.DifficultyLevel{userLevel})
	if err != nil {
		return nil, err
	}

	// 4. Filtration Pipeline (Injuries)
	var filtered []domain.Exercise
	for _, ex := range candidates {
		// Filter By Injury (Contraindications)
		isSafe := true
		for _, contra := range ex.Contraindications {
			for _, injury := range constraints.UserInjuries {
				if contra == injury {
					isSafe = false
					break
				}
			}
		}
		if isSafe {
			filtered = append(filtered, ex)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no exercises found matching constraints")
	}

	// 5. Assembly: Fill Duration
	// Heuristic: Warmup (5m) + Main (X * 5-10m) + Cooldown (5m)
	// Let's simplified: Each exercise ~5 mins (3 sets)
	targetExercises := (constraints.DurationMinutes - 10) / 5
	if targetExercises < 1 {
		targetExercises = 1
	}

	plan := &domain.WorkoutPlan{
		ID:              "", // UUID hooks
		UserID:          userID,
		Date:            time.Now(),
		DurationMinutes: constraints.DurationMinutes,
		Status:          "PENDING",
		EstCaloriesBurn: 0, // Calculate sum later
		Exercises:       []domain.WorkoutItem{},
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	perm := r.Perm(len(filtered))

	for i, idx := range perm {
		if i >= targetExercises {
			break
		}
		ex := filtered[idx]

		plan.Exercises = append(plan.Exercises, domain.WorkoutItem{
			ExerciseID:  ex.ID,
			Exercise:    ex,
			Order:       i + 1,
			Sets:        3,
			Reps:        12, // standard hypertrophy
			RestSec:     60,
			DurationSec: 0, // for cardio
		})
		plan.EstCaloriesBurn += 50 // Mock calorie burn per exercise
	}

	// 6. Save Plan
	err = s.repo.CreateWorkoutPlan(ctx, plan)
	if err != nil {
		return nil, err
	}

	return plan, nil
}
