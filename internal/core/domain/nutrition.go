package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// Enums for Diet Types
type DietType string

const (
	DietTypeNone       DietType = "NONE"
	DietTypeVegan      DietType = "VEGAN"
	DietTypeVegetarian DietType = "VEGETARIAN"
	DietTypeKeto       DietType = "KETO"
	DietTypePaleo      DietType = "PALEO"
)

// FoodPreference stores user-specific nutrition settings
type FoodPreference struct {
	ID                  string         `gorm:"primaryKey;type:text"`
	UserID              string         `gorm:"uniqueIndex;not null;type:text"`
	Allergies           pq.StringArray `gorm:"type:text[]"`
	DislikedIngredients pq.StringArray `gorm:"type:text[]"`
	DietType            DietType       `gorm:"type:varchar(50);default:'NONE'"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// MealSuggestion represents a predefined meal in the catalog
type MealSuggestion struct {
	ID          string         `gorm:"primaryKey;type:text"`
	Name        string         `gorm:"not null;type:text"`
	Calories    float64        `gorm:"not null"`
	Protein     float64        `gorm:"not null"`
	Carbs       float64        `gorm:"not null"`
	Fat         float64        `gorm:"not null"`
	ImageURL    string         `gorm:"type:text"`
	Ingredients pq.StringArray `gorm:"type:text[]"`
	DietTags    pq.StringArray `gorm:"type:text[]"` // e.g. ["VEGAN", "GLUTEN_FREE"]
	Embedding   pgvector.Vector `gorm:"type:vector(384)"`
	CreatedAt   time.Time
}

// MealPlan represents a daily plan generated for a user
type MealPlan struct {
	ID             string    `gorm:"primaryKey;type:text"`
	UserID         string    `gorm:"not null;type:text;index"`
	Date           time.Time `gorm:"not null;type:date;index"`
	TargetCalories float64
	TargetProtein  float64
	TargetCarbs    float64
	TargetFat      float64
	// We use a separate association table or struct for items to allow details
	Meals []MealPlanItem `gorm:"foreignKey:MealPlanID"`
}

type MealType string

const (
	MealTypeBreakfast MealType = "BREAKFAST"
	MealTypeLunch     MealType = "LUNCH"
	MealTypeDinner    MealType = "DINNER"
	MealTypeSnack     MealType = "SNACK"
)

type MealPlanItem struct {
	ID               string `gorm:"primaryKey;type:text"`
	MealPlanID       string `gorm:"not null;type:text;index"`
	MealSuggestionID string `gorm:"not null;type:text"`
	MealSuggestion   MealSuggestion
	MealType         MealType `gorm:"type:varchar(20)"`
	IsEatened        bool     `gorm:"default:false"`
}

// Hooks / Constructors

func NewFoodPreference(userID string) *FoodPreference {
	return &FoodPreference{
		ID:        uuid.NewString(),
		UserID:    userID,
		DietType:  DietTypeNone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (fp *FoodPreference) BeforeCreate(tx *gorm.DB) (err error) {
	if fp.ID == "" {
		fp.ID = uuid.NewString()
	}
	return
}
