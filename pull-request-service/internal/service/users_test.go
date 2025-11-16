package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pull-request-service/internal/models"
	"pull-request-service/internal/service"
	"pull-request-service/internal/service/mocks"
)

func TestUsersService_SetUserActiveStatus(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		mockSetup func(
			trx *mocks.TransactionManager,
			repo *mocks.UsersRepository,
		)
		wantErr bool
		wantRes *models.User
	}{
		{
			name: "success",
			mockSetup: func(trx *mocks.TransactionManager, repo *mocks.UsersRepository) {
				user := &models.User{
					UserID:   "u1",
					Username: "test",
					TeamName: "backend",
					IsActive: false,
				}
				updatedUser := &models.User{
					UserID:   "u1",
					Username: "test",
					TeamName: "backend",
					IsActive: true,
				}

				repo.On("GetUser", mock.Anything, "u1").Return(user, nil).Once()
				repo.On("UpdateUserActiveStatus", mock.Anything, "u1", true).Return(nil).Once()
				repo.On("GetUser", mock.Anything, "u1").Return(updatedUser, nil).Once()

				trx.EXPECT().
					WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
			},
			wantErr: false,
			wantRes: &models.User{
				UserID:   "u1",
				Username: "test",
				TeamName: "backend",
				IsActive: true,
			},
		},

		{
			name: "tx_error",
			mockSetup: func(trx *mocks.TransactionManager, repo *mocks.UsersRepository) {
				trx.EXPECT().
					WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(errors.New("tx fail"))
			},
			wantErr: true,
			wantRes: nil,
		},

		{
			name: "error_getUser_first",
			mockSetup: func(trx *mocks.TransactionManager, repo *mocks.UsersRepository) {
				repo.On("GetUser", mock.Anything, "u1").Return(nil, errors.New("not found")).Once()

				trx.EXPECT().
					WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
			},
			wantErr: true,
			wantRes: nil,
		},

		{
			name: "error_update_status",
			mockSetup: func(trx *mocks.TransactionManager, repo *mocks.UsersRepository) {
				user := &models.User{UserID: "u1", TeamName: "backend", Username: "test", IsActive: false}

				repo.On("GetUser", mock.Anything, "u1").Return(user, nil).Once()
				repo.On("UpdateUserActiveStatus", mock.Anything, "u1", true).Return(errors.New("update error")).Once()

				trx.EXPECT().
					WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
			},
			wantErr: true,
			wantRes: nil,
		},

		{
			name: "error_get_updated_user",
			mockSetup: func(trx *mocks.TransactionManager, repo *mocks.UsersRepository) {
				user := &models.User{UserID: "u1", TeamName: "backend", Username: "test", IsActive: false}

				repo.On("GetUser", mock.Anything, "u1").Return(user, nil).Once()
				repo.On("UpdateUserActiveStatus", mock.Anything, "u1", true).Return(nil).Once()
				repo.On("GetUser", mock.Anything, "u1").Return(nil, errors.New("get error")).Once()

				trx.EXPECT().
					WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
			},
			wantErr: true,
			wantRes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			repo := mocks.NewUsersRepository(t)
			trx := mocks.NewTransactionManager(t)

			svc := service.NewUsersService(repo, nil, trx)

			tt.mockSetup(trx, repo)

			res, err := svc.SetUserActiveStatus(ctx, "u1", true)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantRes, res)
			}
		})
	}
}

func TestUsersService_GetUserReviews(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		retPRs  []*models.PullRequestShort
		retErr  error
		wantErr bool
	}{
		{
			name: "success",
			retPRs: []*models.PullRequestShort{
				{PullRequestID: "pr1", PullRequestName: "Fix bug"},
				{PullRequestID: "pr2", PullRequestName: "Add feature"},
			},
			wantErr: false,
		},
		{
			name:    "repo error",
			retErr:  errors.New("boom"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uRepo := mocks.NewUsersRepository(t)
			rRepo := mocks.NewUserReviewRepository(t)
			tx := mocks.NewTransactionManager(t)

			svc := service.NewUsersService(uRepo, rRepo, tx)

			rRepo.EXPECT().
				GetPRsByReviewer(mock.Anything, "u1").
				Return(tt.retPRs, tt.retErr)

			res, err := svc.GetUserReviews(ctx, "u1")

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.retPRs, res)
			}
		})
	}
}

func TestUsersService_GetReviewsStats(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		ret     []*models.ReviewerStats
		err     error
		wantErr bool
	}{
		{
			name: "success",
			ret: []*models.ReviewerStats{
				{UserID: "u1", ReviewsNumber: 10},
				{UserID: "u2", ReviewsNumber: 5},
			},
		},
		{
			name:    "repo error",
			err:     errors.New("fail"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uRepo := mocks.NewUsersRepository(t)
			rRepo := mocks.NewUserReviewRepository(t)
			tx := mocks.NewTransactionManager(t)

			svc := service.NewUsersService(uRepo, rRepo, tx)

			rRepo.EXPECT().
				GetReviewsStats(mock.Anything).
				Return(tt.ret, tt.err)

			res, err := svc.GetReviewsStats(ctx)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.ret, res)
			}
		})
	}
}
