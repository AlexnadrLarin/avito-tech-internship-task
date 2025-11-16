package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pull-request-service/internal/models"
	"pull-request-service/internal/service"
	mocks "pull-request-service/internal/service/mocks"
)

func expectTx(tx *mocks.TransactionManager) {
	tx.EXPECT().
		WithTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
}

func TestCreatePR(t *testing.T) {
	tests := []struct {
		name  string
		setup func(
			pr *mocks.PullRequestRepository,
			rev *mocks.ReviewRepository,
			team *mocks.TeamInfoRepository,
		)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(pr *mocks.PullRequestRepository, rev *mocks.ReviewRepository, team *mocks.TeamInfoRepository) {
				pr.On("CreatePR", mock.Anything, mock.Anything).Return(nil)

				team.On("GetUserTeam", mock.Anything, "author").Return("teamA", nil)
				team.On("GetActiveTeamMembers", mock.Anything, "teamA", "author").Return([]string{"u1", "u2", "u3"}, nil)

				rev.On("AddReviewer", mock.Anything, "pr1", mock.Anything).Return(nil)

				pr.On("GetPR", mock.Anything, "pr1").Return(&models.PullRequest{
					PullRequestID: "pr1",
					AuthorID:      "author",
					Status:        models.StatusOpen,
				}, nil)
				rev.On("GetPRReviewers", mock.Anything, "pr1").Return([]string{"u1", "u2"}, nil)
			},
			wantErr: false,
		},
		{
			name: "CreatePR error",
			setup: func(pr *mocks.PullRequestRepository, rev *mocks.ReviewRepository, team *mocks.TeamInfoRepository) {
				pr.On("CreatePR", mock.Anything, mock.Anything).Return(errors.New("db err"))
			},
			wantErr: true,
		},
		{
			name: "GetUserTeam error",
			setup: func(pr *mocks.PullRequestRepository, rev *mocks.ReviewRepository, team *mocks.TeamInfoRepository) {
				pr.On("CreatePR", mock.Anything, mock.Anything).Return(nil)
				team.On("GetUserTeam", mock.Anything, "author").Return("", errors.New("no team"))
			},
			wantErr: true,
		},
		{
			name: "no candidates",
			setup: func(pr *mocks.PullRequestRepository, rev *mocks.ReviewRepository, team *mocks.TeamInfoRepository) {
				pr.On("CreatePR", mock.Anything, mock.Anything).Return(nil)

				team.On("GetUserTeam", mock.Anything, "author").Return("teamA", nil)
				team.On("GetActiveTeamMembers", mock.Anything, "teamA", "author").Return([]string{}, nil)

				pr.On("GetPR", mock.Anything, "pr1").Return(&models.PullRequest{
					PullRequestID: "pr1",
					AuthorID:      "author",
					Status:        models.StatusOpen,
				}, nil)

				rev.On("GetPRReviewers", mock.Anything, "pr1").Return([]string{}, nil)
			},
			wantErr: false,
		},
		{
			name: "AddReviewer fails",
			setup: func(pr *mocks.PullRequestRepository, rev *mocks.ReviewRepository, team *mocks.TeamInfoRepository) {
				pr.On("CreatePR", mock.Anything, mock.Anything).Return(nil)
				team.On("GetUserTeam", mock.Anything, "author").Return("teamA", nil)
				team.On("GetActiveTeamMembers", mock.Anything, "teamA", "author").Return([]string{"u1"}, nil)

				rev.On("AddReviewer", mock.Anything, "pr1", "u1").Return(errors.New("fail"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			prRepo := mocks.NewPullRequestRepository(t)
			revRepo := mocks.NewReviewRepository(t)
			teamRepo := mocks.NewTeamInfoRepository(t)
			txMgr := mocks.NewTransactionManager(t)

			expectTx(txMgr)
			tt.setup(prRepo, revRepo, teamRepo)

			svc := service.NewPullRequestService(prRepo, revRepo, teamRepo, txMgr)

			pr := &models.PullRequest{
				PullRequestID: "pr1",
				AuthorID:      "author",
			}

			_, err := svc.CreatePR(context.Background(), pr)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMergePR(t *testing.T) {
	prRepo := mocks.NewPullRequestRepository(t)
	revRepo := mocks.NewReviewRepository(t)
	teamRepo := mocks.NewTeamInfoRepository(t)
	txMgr := mocks.NewTransactionManager(t)

	expectTx(txMgr)

	prRepo.On("MergePR", mock.Anything, "pr1").Return(nil)
	prRepo.On("GetPR", mock.Anything, "pr1").Return(&models.PullRequest{
		PullRequestID: "pr1",
		Assigned:      []string{"u1"},
	}, nil)
	revRepo.On("GetPRReviewers", mock.Anything, "pr1").Return([]string{"u1"}, nil)

	svc := service.NewPullRequestService(prRepo, revRepo, teamRepo, txMgr)

	_, err := svc.MergePR(context.Background(), "pr1")
	require.NoError(t, err)
}

func TestMergePR_Error(t *testing.T) {
	prRepo := mocks.NewPullRequestRepository(t)
	revRepo := mocks.NewReviewRepository(t)
	teamRepo := mocks.NewTeamInfoRepository(t)
	txMgr := mocks.NewTransactionManager(t)

	expectTx(txMgr)

	prRepo.On("MergePR", mock.Anything, "pr1").Return(errors.New("fail"))

	svc := service.NewPullRequestService(prRepo, revRepo, teamRepo, txMgr)

	_, err := svc.MergePR(context.Background(), "pr1")
	require.Error(t, err)
}

func TestGetPR(t *testing.T) {
	tests := []struct {
		name  string
		setup func(
			pr *mocks.PullRequestRepository,
			rev *mocks.ReviewRepository,
		)
		wantErr bool
		wantPR  *models.PullRequest
	}{
		{
			name: "success",
			setup: func(pr *mocks.PullRequestRepository, rev *mocks.ReviewRepository) {
				pr.On("GetPR", mock.Anything, "pr1").Return(&models.PullRequest{
					PullRequestID:   "pr1",
					PullRequestName: "test PR",
					AuthorID:        "author1",
					Status:          models.StatusOpen,
				}, nil)
				rev.On("GetPRReviewers", mock.Anything, "pr1").Return([]string{"u1", "u2"}, nil)
			},
			wantErr: false,
			wantPR: &models.PullRequest{
				PullRequestID:   "pr1",
				PullRequestName: "test PR",
				AuthorID:        "author1",
				Status:          models.StatusOpen,
				Assigned:        []string{"u1", "u2"},
			},
		},
		{
			name: "GetPR error",
			setup: func(pr *mocks.PullRequestRepository, rev *mocks.ReviewRepository) {
				pr.On("GetPR", mock.Anything, "pr1").Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "GetPRReviewers error",
			setup: func(pr *mocks.PullRequestRepository, rev *mocks.ReviewRepository) {
				pr.On("GetPR", mock.Anything, "pr1").Return(&models.PullRequest{
					PullRequestID: "pr1",
					Status:        models.StatusOpen,
				}, nil)
				rev.On("GetPRReviewers", mock.Anything, "pr1").Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "no reviewers",
			setup: func(pr *mocks.PullRequestRepository, rev *mocks.ReviewRepository) {
				pr.On("GetPR", mock.Anything, "pr1").Return(&models.PullRequest{
					PullRequestID:   "pr1",
					PullRequestName: "test PR",
					AuthorID:        "author1",
					Status:          models.StatusOpen,
				}, nil)
				rev.On("GetPRReviewers", mock.Anything, "pr1").Return([]string{}, nil)
			},
			wantErr: false,
			wantPR: &models.PullRequest{
				PullRequestID:   "pr1",
				PullRequestName: "test PR",
				AuthorID:        "author1",
				Status:          models.StatusOpen,
				Assigned:        []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := mocks.NewPullRequestRepository(t)
			revRepo := mocks.NewReviewRepository(t)
			teamRepo := mocks.NewTeamInfoRepository(t)
			txMgr := mocks.NewTransactionManager(t)

			tt.setup(prRepo, revRepo)

			svc := service.NewPullRequestService(prRepo, revRepo, teamRepo, txMgr)

			result, err := svc.GetPR(context.Background(), "pr1")

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantPR.PullRequestID, result.PullRequestID)
				require.Equal(t, tt.wantPR.PullRequestName, result.PullRequestName)
				require.Equal(t, tt.wantPR.AuthorID, result.AuthorID)
				require.Equal(t, tt.wantPR.Status, result.Status)
				require.Equal(t, tt.wantPR.Assigned, result.Assigned)
			}
		})
	}
}

func TestReassignReviewer(t *testing.T) {
	tests := []struct {
		name  string
		setup func(
			pr *mocks.PullRequestRepository,
			rev *mocks.ReviewRepository,
			team *mocks.TeamInfoRepository,
		)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(pr *mocks.PullRequestRepository, rev *mocks.ReviewRepository, team *mocks.TeamInfoRepository) {
				pr.On("GetPR", mock.Anything, "pr1").Return(&models.PullRequest{
					PullRequestID: "pr1",
					AuthorID:      "author",
					Status:        models.StatusOpen,
					Assigned:      []string{"old"},
				}, nil)
				rev.On("GetPRReviewers", mock.Anything, "pr1").Return([]string{"old"}, nil)

				team.On("GetUserTeam", mock.Anything, "old").Return("teamA", nil)
				team.On("GetActiveTeamMembers", mock.Anything, "teamA", "old").Return([]string{"c1"}, nil)

				rev.On("RemoveReviewer", mock.Anything, "pr1", "old").Return(nil)
				rev.On("AddReviewer", mock.Anything, "pr1", "c1").Return(nil)

				pr.On("GetPR", mock.Anything, "pr1").Return(&models.PullRequest{
					PullRequestID: "pr1",
					AuthorID:      "author",
					Status:        models.StatusOpen,
					Assigned:      []string{"c1"},
				}, nil)
				rev.On("GetPRReviewers", mock.Anything, "pr1").Return([]string{"c1"}, nil)
			},
			wantErr: false,
		},
		{
			name: "PR merged",
			setup: func(pr *mocks.PullRequestRepository, rev *mocks.ReviewRepository, team *mocks.TeamInfoRepository) {
				pr.On("GetPR", mock.Anything, "pr1").Return(&models.PullRequest{
					Status:   models.StatusMerged,
					Assigned: []string{"old"},
				}, nil)
				rev.On("GetPRReviewers", mock.Anything, "pr1").Return([]string{"old"}, nil)
			},
			wantErr: true,
		},
		{
			name: "old reviewer not assigned",
			setup: func(pr *mocks.PullRequestRepository, rev *mocks.ReviewRepository, team *mocks.TeamInfoRepository) {
				pr.On("GetPR", mock.Anything, "pr1").Return(&models.PullRequest{
					Status:   models.StatusOpen,
					Assigned: []string{"x"},
				}, nil)
				rev.On("GetPRReviewers", mock.Anything, "pr1").Return([]string{"x"}, nil)
			},
			wantErr: true,
		},
		{
			name: "no candidates",
			setup: func(pr *mocks.PullRequestRepository, rev *mocks.ReviewRepository, team *mocks.TeamInfoRepository) {
				pr.On("GetPR", mock.Anything, "pr1").Return(&models.PullRequest{
					Status:   models.StatusOpen,
					AuthorID: "author",
					Assigned: []string{"old"},
				}, nil)
				rev.On("GetPRReviewers", mock.Anything, "pr1").Return([]string{"old"}, nil)

				team.On("GetUserTeam", mock.Anything, "old").Return("teamA", nil)
				team.On("GetActiveTeamMembers", mock.Anything, "teamA", "old").Return([]string{}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			prRepo := mocks.NewPullRequestRepository(t)
			revRepo := mocks.NewReviewRepository(t)
			teamRepo := mocks.NewTeamInfoRepository(t)
			txMgr := mocks.NewTransactionManager(t)

			expectTx(txMgr)
			tt.setup(prRepo, revRepo, teamRepo)

			svc := service.NewPullRequestService(prRepo, revRepo, teamRepo, txMgr)

			_, _, err := svc.ReassignReviewer(context.Background(), "pr1", "old")

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
