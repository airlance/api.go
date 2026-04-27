package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/resoul/api/internal/domain"
)

type accountService struct {
	repo          domain.AccountRepository
	storageClient *minio.Client
}

// NewAccountService returns an AccountService backed by the given repository and storage client.
func NewAccountService(repo domain.AccountRepository, storageClient *minio.Client) domain.AccountService {
	return &accountService{repo: repo, storageClient: storageClient}
}

// GetOrCreate returns the existing account for userID.
// If no account exists yet it creates and persists an empty one with a unique bucket.
func (s *accountService) GetOrCreate(ctx context.Context, userID string) (*domain.Account, error) {
	account, err := s.repo.FindByUserID(ctx, userID)
	if err == nil {
		return account, nil
	}

	if err != domain.ErrNotFound {
		return nil, err
	}

	// No account yet — create an empty one with a unique bucket.
	bucketName := fmt.Sprintf("user-%s", uuid.New().String())
	err = s.storageClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		return nil, fmt.Errorf("create bucket: %w", err)
	}

	blank := &domain.Account{UserID: userID, BucketName: bucketName}
	return s.repo.Upsert(ctx, blank)
}

// Update applies non-nil fields from inp to the account owned by userID.
// Returns domain.ErrInvalidInput when inp carries no changes.
// Returns domain.ErrNotFound when no account exists for userID.
func (s *accountService) Update(ctx context.Context, userID string, inp domain.UpdateAccountInput) (*domain.Account, error) {
	if inp.DisplayName == nil && inp.AvatarURL == nil && inp.Bio == nil {
		return nil, domain.ErrInvalidInput
	}

	account, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err // propagates ErrNotFound as-is
	}

	if inp.DisplayName != nil {
		account.DisplayName = *inp.DisplayName
	}

	if inp.AvatarURL != nil {
		account.AvatarURL = *inp.AvatarURL
	}

	if inp.Bio != nil {
		account.Bio = *inp.Bio
	}

	return s.repo.Upsert(ctx, account)
}
