package domain

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionStatus string

const (
	SubscriptionStatusFree    SubscriptionStatus = "FREE"
	SubscriptionStatusPremium SubscriptionStatus = "PREMIUM"
)

// User struct adapted to provided Drizzle schema + Business fields
type User struct {
	ID            string             `gorm:"primaryKey;type:text"`
	Name          string             `gorm:"not null;type:text"`
	Email         string             `gorm:"uniqueIndex;not null;type:text"`
	EmailVerified bool               `gorm:"not null"`
	Image         *string            `gorm:"type:text"`
	CreatedAt     time.Time          `gorm:"not null"`
	UpdatedAt     time.Time          `gorm:"not null"`
	
	// Business fields
	SubscriptionStatus SubscriptionStatus `gorm:"type:varchar(20);default:'FREE'"`
	Weight             float64
	Height             float64
}

type Session struct {
	ID        string    `gorm:"primaryKey;type:text"`
	ExpiresAt time.Time `gorm:"not null"`
	Token     string    `gorm:"uniqueIndex;not null;type:text"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
	IpAddress *string   `gorm:"type:text"`
	UserAgent *string   `gorm:"type:text"`
	UserId    string    `gorm:"not null;type:text"`
}

type Account struct {
	ID                    string    `gorm:"primaryKey;type:text"`
	AccountId             string    `gorm:"not null;type:text"`
	ProviderId            string    `gorm:"not null;type:text"`
	UserId                string    `gorm:"not null;type:text"`
	AccessToken           *string   `gorm:"type:text"`
	RefreshToken          *string   `gorm:"type:text"`
	IdToken               *string   `gorm:"type:text"`
	AccessTokenExpiresAt  *time.Time
	RefreshTokenExpiresAt *time.Time
	Scope                 *string   `gorm:"type:text"`
	Password              *string   `gorm:"type:text"`
	CreatedAt             time.Time `gorm:"not null"`
	UpdatedAt             time.Time `gorm:"not null"`
}

type Verification struct {
	ID         string    `gorm:"primaryKey;type:text"`
	Identifier string    `gorm:"not null;type:text"`
	Value      string    `gorm:"not null;type:text"`
	ExpiresAt  time.Time `gorm:"not null"`
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

func NewUser(email, name string) *User {
	now := time.Now()
	return &User{
		ID:                 uuid.NewString(), // Generate UUID as string
		Name:               name,
		Email:              email,
		EmailVerified:      false,
		SubscriptionStatus: SubscriptionStatusFree,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}
