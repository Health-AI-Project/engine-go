package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type SubscriptionStatus string

const (
	SubscriptionStatusFree    SubscriptionStatus = "FREE"
	SubscriptionStatusPremium SubscriptionStatus = "PREMIUM"
)

// User struct adapted to provided Drizzle schema + Business fields
type User struct {
	ID            string    `gorm:"primaryKey;column:id;type:text"`
	Name          string    `gorm:"not null;column:name;type:text"`
	Email         string    `gorm:"uniqueIndex:user_email_unique;not null;column:email;type:text"`
	EmailVerified bool      `gorm:"not null;column:email_verified"`
	Image         *string   `gorm:"type:text;column:image"`
	CreatedAt     time.Time `gorm:"not null;column:created_at"`
	UpdatedAt     time.Time `gorm:"not null;column:updated_at"`

	// Business fields
	SubscriptionStatus SubscriptionStatus `gorm:"type:varchar(20);default:'FREE';column:subscriptionStatus"`
	Weight             float64            `gorm:"column:weight"`
	Height             float64            `gorm:"column:height"`
	DateOfBirth        *time.Time         `gorm:"column:dateOfBirth"`
	HealthProfile      *HealthProfile     `gorm:"foreignKey:UserID;references:ID"`
}

type HealthProfile struct {
	ID        string         `gorm:"primaryKey;type:text"`
	UserID    string         `gorm:"uniqueIndex;not null;type:text"`
	Goals     pq.StringArray `gorm:"type:text[]"`
	Allergies pq.StringArray `gorm:"type:text[]"`
	CreatedAt time.Time      `gorm:"column:createdAt"`
	UpdatedAt time.Time      `gorm:"column:updatedAt"`
}

func (HealthProfile) TableName() string {
	return "health_profile"
}

type Session struct {
	ID        string    `gorm:"primaryKey;column:id;type:text"`
	ExpiresAt time.Time `gorm:"not null;column:expiresAt"`
	Token     string    `gorm:"uniqueIndex;not null;column:token;type:text"`
	CreatedAt time.Time `gorm:"not null;column:createdAt"`
	UpdatedAt time.Time `gorm:"not null;column:updatedAt"`
	IpAddress *string   `gorm:"type:text;column:ipAddress"`
	UserAgent *string   `gorm:"type:text;column:userAgent"`
	UserId    string    `gorm:"not null;column:userId;type:text"`
}

func (Session) TableName() string {
	return "session"
}

type Account struct {
	ID                    string     `gorm:"primaryKey;column:id;type:text"`
	AccountId             string     `gorm:"not null;column:accountId;type:text"`
	ProviderId            string     `gorm:"not null;column:providerId;type:text"`
	UserId                string     `gorm:"not null;column:userId;type:text"`
	AccessToken           *string    `gorm:"type:text;column:accessToken"`
	RefreshToken          *string    `gorm:"type:text;column:refreshToken"`
	IdToken               *string    `gorm:"type:text;column:idToken"`
	AccessTokenExpiresAt  *time.Time `gorm:"column:accessTokenExpiresAt"`
	RefreshTokenExpiresAt *time.Time `gorm:"column:refreshTokenExpiresAt"`
	Scope                 *string    `gorm:"type:text;column:scope"`
	Password              *string    `gorm:"type:text;column:password"`
	CreatedAt             time.Time  `gorm:"not null;column:createdAt"`
	UpdatedAt             time.Time  `gorm:"not null;column:updatedAt"`
}

func (Account) TableName() string {
	return "account"
}

type Verification struct {
	ID         string     `gorm:"primaryKey;column:id;type:text"`
	Identifier string     `gorm:"not null;column:identifier;type:text"`
	Value      string     `gorm:"not null;column:value;type:text"`
	ExpiresAt  time.Time  `gorm:"not null;column:expiresAt"`
	CreatedAt  *time.Time `gorm:"column:createdAt"`
	UpdatedAt  *time.Time `gorm:"column:updatedAt"`
}

func (Verification) TableName() string {
	return "verification"
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

// TableName overrides the table name used by User to `user`
func (User) TableName() string {
	return "user"
}
