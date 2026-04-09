package services

import (
	"context"
	"time"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/ports"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserService struct {
	repo ports.UserRepository
}

func NewUserService(repo ports.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) UpdateBiometrics(ctx context.Context, id string, weight, height float64) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	user.Weight = weight
	user.Height = height

	return s.repo.Update(ctx, user)
}

func (s *UserService) UpdateHealthProfile(ctx context.Context, id string, dob *time.Time, goals, allergies []string, weight, height float64) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	user.DateOfBirth = dob
	user.Weight = weight
	user.Height = height

	if user.HealthProfile == nil {
		user.HealthProfile = &domain.HealthProfile{
			ID:        uuid.NewString(),
			UserID:    user.ID,
			Goals:     pq.StringArray(goals),
			Allergies: pq.StringArray(allergies),
		}
	} else {
		user.HealthProfile.Goals = pq.StringArray(goals)
		user.HealthProfile.Allergies = pq.StringArray(allergies)
	}

	return s.repo.Update(ctx, user)
}

// CanAccessAdvancedFeatures checks if the user has PREMIUM status.
// This serves as the "Freemium Guard" for future features.
func (s *UserService) CanAccessAdvancedFeatures(user *domain.User) bool {
	return user.SubscriptionStatus == domain.SubscriptionStatusPremium
}
