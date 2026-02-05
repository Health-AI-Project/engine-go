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

func TestCanAccessAdvancedFeatures(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	service := services.NewUserService(mockRepo)

	t.Run("User FREE -> Should return false", func(t *testing.T) {
		user := &domain.User{
			SubscriptionStatus: domain.SubscriptionStatusFree,
		}
		assert.False(t, service.CanAccessAdvancedFeatures(user))
	})

	t.Run("User PREMIUM -> Should return true", func(t *testing.T) {
		user := &domain.User{
			SubscriptionStatus: domain.SubscriptionStatusPremium,
		}
		assert.True(t, service.CanAccessAdvancedFeatures(user))
	})
}

func TestGetUser(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	service := services.NewUserService(mockRepo)

	t.Run("Found", func(t *testing.T) {
		user := &domain.User{ID: "123"}
		mockRepo.On("GetByID", context.Background(), "123").Return(user, nil).Once()

		res, err := service.GetUser(context.Background(), "123")
		assert.NoError(t, err)
		assert.Equal(t, user, res)
	})
}

func TestUpdateBiometrics(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	service := services.NewUserService(mockRepo)

	t.Run("Update OK", func(t *testing.T) {
		user := &domain.User{ID: "123"}
		mockRepo.On("GetByID", context.Background(), "123").Return(user, nil).Once()
		mockRepo.On("Update", context.Background(), mock.MatchedBy(func(u *domain.User) bool {
			return u.Weight == 80 && u.Height == 180
		})).Return(nil).Once()

		err := service.UpdateBiometrics(context.Background(), "123", 80, 180)
		assert.NoError(t, err)
	})
}
func TestUpdateHealthProfile(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	service := services.NewUserService(mockRepo)

	t.Run("Update Health Profile OK", func(t *testing.T) {
		user := &domain.User{ID: "123"}
		mockRepo.On("GetByID", context.Background(), "123").Return(user, nil).Once()
		mockRepo.On("Update", context.Background(), mock.MatchedBy(func(u *domain.User) bool {
			return u.DateOfBirth != nil &&
				u.Weight == 80.0 &&
				u.HealthProfile != nil &&
				u.HealthProfile.Goals[0] == "perdre du poids" &&
				u.HealthProfile.Allergies[0] == "arachides"
		})).Return(nil).Once()

		dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
		err := service.UpdateHealthProfile(context.Background(), "123", &dob, []string{"perdre du poids"}, []string{"arachides"}, 80.0)
		assert.NoError(t, err)
	})
}
