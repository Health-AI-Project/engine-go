package mocks

import (
	"context"
	"time"

	"healthai/engine/internal/core/domain"

	"github.com/stretchr/testify/mock"
)

type MockActivityRepository struct {
	mock.Mock
}

func (m *MockActivityRepository) GetDailyLogByDate(ctx context.Context, userID string, date time.Time) (*domain.DailyLog, error) {
	args := m.Called(ctx, userID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DailyLog), args.Error(1)
}

func (m *MockActivityRepository) GetDailyLogHistory(ctx context.Context, userID string, days int) ([]domain.DailyLog, error) {
	args := m.Called(ctx, userID, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DailyLog), args.Error(1)
}

func (m *MockActivityRepository) CreateDailyLog(ctx context.Context, log *domain.DailyLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockActivityRepository) UpdateDailyLog(ctx context.Context, log *domain.DailyLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockActivityRepository) CreateWorkout(ctx context.Context, workout *domain.Workout) error {
	args := m.Called(ctx, workout)
	return args.Error(0)
}
