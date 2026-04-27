package domain

import (
	"context"
	"time"
)

// Account is the application-level user account.
// It is distinct from auth.users — it stores product data owned by this service.
type Account struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      string    `gorm:"type:uuid;uniqueIndex;not null"`
	BucketName  string    `gorm:"type:text;uniqueIndex;not null"`
	DisplayName string    `gorm:"type:text"`
	AvatarURL   string    `gorm:"type:text"`
	Bio         string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// UpdateAccountInput carries the fields a user may change on their account.
// Pointer fields allow partial updates — nil means "not provided, do not overwrite".
type UpdateAccountInput struct {
	DisplayName *string `json:"display_name" binding:"omitempty,max=100"`
	AvatarURL   *string `json:"avatar_url"   binding:"omitempty,url,max=2048"`
	Bio         *string `json:"bio"          binding:"omitempty,max=500"`
}

// AccountRepository is the persistence port for Account.
// Implementations live in internal/infrastructure/db/.
type AccountRepository interface {
	// FindByUserID returns the account for the given auth user ID.
	// Returns ErrNotFound when no account exists yet.
	FindByUserID(ctx context.Context, userID string) (*Account, error)

	// Upsert creates the account if it does not exist, otherwise updates it.
	Upsert(ctx context.Context, account *Account) (*Account, error)
}

// AccountService is the business-logic port for account operations.
// Implementations live in internal/service/.
type AccountService interface {
	// GetOrCreate returns the existing account for userID, or creates a new empty one.
	GetOrCreate(ctx context.Context, userID string) (*Account, error)

	// Update applies inp to the account owned by userID and persists the result.
	// Returns ErrInvalidInput if inp carries no changes.
	Update(ctx context.Context, userID string, inp UpdateAccountInput) (*Account, error)
}
