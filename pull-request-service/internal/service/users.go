package service

import (
	"context"
	"fmt"

	"pull-request-service/internal/models"
)

type UsersRepository interface {
	UpdateUserActiveStatus(ctx context.Context, userID string, isActive bool) error
	GetUser(ctx context.Context, userID string) (*models.User, error)
	UpsertUser(ctx context.Context, userID, username, teamName string, isActive bool) error
}

type UserReviewRepository interface {
	GetPRsByReviewer(ctx context.Context, userID string) ([]*models.PullRequestShort, error)
	GetReviewsStats(ctx context.Context) ([]*models.ReviewerStats, error)
}

type UsersService struct {
	usersRepo  UsersRepository
	reviewRepo UserReviewRepository
	txMgr      TransactionManager
}

func NewUsersService(u UsersRepository, r UserReviewRepository, txMgr TransactionManager) *UsersService {
	return &UsersService{
		usersRepo:  u,
		reviewRepo: r,
		txMgr:      txMgr,
	}
}

func (s *UsersService) SetUserActiveStatus(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	var result *models.User

	err := s.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		if _, err := s.usersRepo.GetUser(txCtx, userID); err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		if err := s.usersRepo.UpdateUserActiveStatus(txCtx, userID, isActive); err != nil {
			return fmt.Errorf("updating user status: %w", err)
		}

		var err error
		result, err = s.usersRepo.GetUser(txCtx, userID)
		if err != nil {
			return fmt.Errorf("getting updated user: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *UsersService) GetUserReviews(ctx context.Context, userID string) ([]*models.PullRequestShort, error) {
	prs, err := s.reviewRepo.GetPRsByReviewer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user reviews: %w", err)
	}
	return prs, nil
}

func (s *UsersService) GetReviewsStats(ctx context.Context) ([]*models.ReviewerStats, error) {
	return s.reviewRepo.GetReviewsStats(ctx)
}

