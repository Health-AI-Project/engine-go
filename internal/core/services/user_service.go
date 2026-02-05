package services

import (
	"context"

	"healthai/engine/internal/core/domain"
	"healthai/engine/internal/core/ports"
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
	
	// Example of business logic potentially using CanAccessAdvancedFeatures check here if needed in future
	// if !s.CanAccessAdvancedFeatures(user) { ... }

	return s.repo.Update(ctx, user)
}

// CanAccessAdvancedFeatures checks if the user has PREMIUM status.
// This serves as the "Freemium Guard" for future features.
func (s *UserService) CanAccessAdvancedFeatures(user *domain.User) bool {
	return user.SubscriptionStatus == domain.SubscriptionStatusPremium
}
