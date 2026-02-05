package domain

import (
	"time"

	"github.com/google/uuid"
)

type WorkoutType string

const (
	WorkoutTypeCardio WorkoutType = "CARDIO"
	WorkoutTypeStrength WorkoutType = "STRENGTH"
)

type DailyLog struct {
	ID            string    `gorm:"primaryKey;type:text"`
	UserID        string    `gorm:"not null;type:text;index"` // Indexed for quick lookup by user+date
	Date          time.Time `gorm:"not null;type:date;index"` // Date part only usually
	TotalCalories float64
	TotalProtein  float64
	TotalCarbs    float64
	TotalFat      float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Workout struct {
	ID             string      `gorm:"primaryKey;type:text"`
	UserID         string      `gorm:"not null;type:text"` // Direct link to user
	DailyLogID     *string     `gorm:"type:text"` // Optional link to daily log
	Type           WorkoutType `gorm:"type:varchar(50)"`
	Duration       int         // Minutes
	CaloriesBurned float64
	Date           time.Time   `gorm:"not null"`
	CreatedAt      time.Time
}

func NewDailyLog(userID string, date time.Time) *DailyLog {
	return &DailyLog{
		ID:        uuid.NewString(), // Keeping string ID convention
		UserID:    userID,
		Date:      date,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func NewWorkout(userID string, wType WorkoutType, duration int, calories float64) *Workout {
	return &Workout{
		ID:             uuid.NewString(),
		UserID:         userID,
		Type:           wType,
		Duration:       duration,
		CaloriesBurned: calories,
		Date:           time.Now(),
		CreatedAt:      time.Now(),
	}
}
