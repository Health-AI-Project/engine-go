package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type ExerciseType string

const (
	ExerciseTypeCardio      ExerciseType = "CARDIO"
	ExerciseTypeStrength    ExerciseType = "STRENGTH"
	ExerciseTypeFlexibility ExerciseType = "FLEXIBILITY"
)

type EquipmentType string

const (
	EquipmentNone      EquipmentType = "NONE"
	EquipmentDumbbells EquipmentType = "DUMBBELLS"
	EquipmentBarbell   EquipmentType = "BARBELL"
	EquipmentMachine   EquipmentType = "MACHINE"
	EquipmentBand      EquipmentType = "BAND"
)

type DifficultyLevel string

const (
	DifficultyBeginner     DifficultyLevel = "BEGINNER"
	DifficultyIntermediate DifficultyLevel = "INTERMEDIATE"
	DifficultyAdvanced     DifficultyLevel = "ADVANCED"
)

// Exercise catalog
type Exercise struct {
	ID                string          `gorm:"primaryKey;type:text"`
	Name              string          `gorm:"not null;type:text"`
	Type              ExerciseType    `gorm:"type:varchar(50)"`
	RequiredEquipment EquipmentType   `gorm:"type:varchar(50)"`
	Difficulty        DifficultyLevel `gorm:"type:varchar(50)"`
	MusclesTargeted   pq.StringArray  `gorm:"type:text[]"` // e.g. ["Chest", "Triceps"]
	Contraindications pq.StringArray  `gorm:"type:text[]"` // e.g. ["ShoulderInjury"]
	ImageURL          string          `gorm:"type:text"`
	VideoURL          string          `gorm:"type:text"`
	CreatedAt         time.Time
}

// WorkoutPlan generated for a user
type WorkoutPlan struct {
	ID              string    `gorm:"primaryKey;type:text"`
	UserID          string    `gorm:"not null;type:text;index"`
	Date            time.Time `gorm:"not null;type:date;index"`
	DurationMinutes int
	EstCaloriesBurn float64
	Status          string        `gorm:"type:varchar(20);default:'PENDING'"` // PENDING, COMPLETED, SKIPPED
	Exercises       []WorkoutItem `gorm:"foreignKey:WorkoutPlanID"`
}

type WorkoutItem struct {
	ID            string `gorm:"primaryKey;type:text"`
	WorkoutPlanID string `gorm:"not null;type:text;index"`
	ExerciseID    string `gorm:"not null;type:text"`
	Exercise      Exercise
	Order         int // 1, 2, 3...
	Sets          int
	Reps          int
	DurationSec   int // For cardio or timed exercises
	RestSec       int
}

// Constructors

func (e *Exercise) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	return
}

func (wp *WorkoutPlan) BeforeCreate(tx *gorm.DB) (err error) {
	if wp.ID == "" {
		wp.ID = uuid.NewString()
	}
	return
}
