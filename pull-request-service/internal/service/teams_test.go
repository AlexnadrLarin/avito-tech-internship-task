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

func TestTeamsService_AddTeam(t *testing.T) {
	type fields struct {
		teamsRepoErr error
		userErrAt    int
		txErr        error
	}

	tests := []struct {
		name          string
		fields        fields
		expectInsert  bool
		expectUpserts int
		wantErr       bool
	}{
		{
			name:          "success",
			fields:        fields{userErrAt: -1},
			expectInsert:  true,
			expectUpserts: 2,
			wantErr:       false,
		},
		{
			name: "error InsertTeam",
			fields: fields{
				teamsRepoErr: errors.New("insert error"),
			},
			expectInsert:  true,
			expectUpserts: 0,
			wantErr:       true,
		},
		{
			name: "error UpsertUser",
			fields: fields{
				userErrAt: 1,
			},
			expectInsert:  true,
			expectUpserts: 2,
			wantErr:       true,
		},
		{
			name: "transaction manager returns error",
			fields: fields{
				txErr: errors.New("tx error"),
			},
			expectInsert:  false,
			expectUpserts: 0,
			wantErr:       true,
		},
	}

	team := &models.Team{
		TeamName: "backend",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "alice", IsActive: true},
			{UserID: "u2", Username: "bob", IsActive: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			teamsRepo := mocks.NewTeamsRepository(t)
			usersRepo := mocks.NewUsersRepository(t)
			txMgr := mocks.NewTransactionManager(t)

			txMgr.EXPECT().
				WithTransaction(ctx, mock.Anything).
				RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					if tt.fields.txErr != nil {
						return tt.fields.txErr
					}
					return fn(ctx)
				})

			if tt.expectInsert {
				teamsRepo.EXPECT().
					InsertTeam(mock.Anything, "backend").
					Return(tt.fields.teamsRepoErr)
			}

			for i, m := range team.Members {
				if i >= tt.expectUpserts {
					break
				}

				var err error
				if tt.fields.userErrAt == i {
					err = errors.New("upsert error")
				}

				usersRepo.EXPECT().
					UpsertUser(
						mock.Anything,
						m.UserID,
						m.Username,
						"backend",
						m.IsActive,
					).
					Return(err)
			}

			svc := service.NewTeamService(teamsRepo, usersRepo, txMgr)

			err := svc.AddTeam(ctx, team)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTeamsService_GetTeam(t *testing.T) {
	tests := []struct {
		name    string
		retTeam *models.Team
		retErr  error
		wantErr bool
	}{
		{
			name: "success",
			retTeam: &models.Team{
				TeamName: "backend",
			},
			retErr:  nil,
			wantErr: false,
		},
		{
			name:    "repo error",
			retTeam: nil,
			retErr:  errors.New("boom"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			teamsRepo := mocks.NewTeamsRepository(t)
			usersRepo := mocks.NewUsersRepository(t)
			txMgr := mocks.NewTransactionManager(t)

			teamsRepo.EXPECT().
				GetTeam(mock.Anything, "backend").
				Return(tt.retTeam, tt.retErr)

			svc := service.NewTeamService(teamsRepo, usersRepo, txMgr)

			team, err := svc.GetTeam(context.Background(), "backend")
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, team)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.retTeam, team)
			}
		})
	}
}

func TestTeamsService_GetUserTeam(t *testing.T) {
	tests := []struct {
		name    string
		retName string
		retErr  error
		wantErr bool
	}{
		{
			name:    "success",
			retName: "backend",
			retErr:  nil,
		},
		{
			name:    "repo error",
			retName: "",
			retErr:  errors.New("fail"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teamsRepo := mocks.NewTeamsRepository(t)
			usersRepo := mocks.NewUsersRepository(t)
			txMgr := mocks.NewTransactionManager(t)

			teamsRepo.EXPECT().
				GetUserTeam(mock.Anything, "u1").
				Return(tt.retName, tt.retErr)

			svc := service.NewTeamService(teamsRepo, usersRepo, txMgr)

			team, err := svc.GetUserTeam(context.Background(), "u1")
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, team)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.retName, team)
			}
		})
	}
}

func TestTeamsService_GetActiveTeamMembers(t *testing.T) {
	tests := []struct {
		name       string
		retMembers []string
		retErr     error
		wantErr    bool
	}{
		{
			name:       "success",
			retMembers: []string{"u1", "u2"},
			retErr:     nil,
		},
		{
			name:       "repo error",
			retMembers: nil,
			retErr:     errors.New("repo error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teamsRepo := mocks.NewTeamsRepository(t)
			usersRepo := mocks.NewUsersRepository(t)
			txMgr := mocks.NewTransactionManager(t)

			teamsRepo.EXPECT().
				GetActiveTeamMembers(mock.Anything, "backend", "exclude").
				Return(tt.retMembers, tt.retErr)

			svc := service.NewTeamService(teamsRepo, usersRepo, txMgr)

			res, err := svc.GetActiveTeamMembers(context.Background(), "backend", "exclude")
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.retMembers, res)
			}
		})
	}
}
