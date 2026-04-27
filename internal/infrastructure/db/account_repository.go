package db

import (
	"context"
	"errors"

	"github.com/resoul/api/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type accountRepository struct {
	db *gorm.DB
}

// NewAccountRepository returns a GORM-backed AccountRepository.
func NewAccountRepository(db *gorm.DB) domain.AccountRepository {
	return &accountRepository{db: db}
}

// FindByUserID returns the account for the given auth user ID.
// Returns domain.ErrNotFound when no account exists yet.
func (r *accountRepository) FindByUserID(ctx context.Context, userID string) (*domain.Account, error) {
	var a domain.Account

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&a).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &a, nil
}

// Upsert inserts a new account or updates the existing one matched by user_id.
// Only non-zero fields from account are written on conflict.
func (r *accountRepository) Upsert(ctx context.Context, account *domain.Account) (*domain.Account, error) {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"display_name",
				"avatar_url",
				"bio",
				"updated_at",
			}),
		}).
		Create(account).Error

	if err != nil {
		return nil, err
	}

	return account, nil
}
