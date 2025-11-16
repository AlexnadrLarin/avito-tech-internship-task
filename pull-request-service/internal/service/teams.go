package service

import (
	"context"
	"fmt"

	"pull-request-service/internal/models"
)

type TeamsRepository interface {
	InsertTeam(ctx context.Context, teamName string) error
	GetTeam(ctx context.Context, teamName string) (*models.Team, error)
	GetUserTeam(ctx context.Context, userID string) (string, error)
	GetActiveTeamMembers(ctx context.Context, teamName string, excludeUserID string) ([]string, error)
}

type TeamUsersRepository interface {
	UpsertUser(ctx context.Context, userID, username, teamName string, isActive bool) error
}

type TeamsService struct {
	teamsRepo TeamsRepository
	usersRepo TeamUsersRepository
	txMgr     TransactionManager
}

func NewTeamService(t TeamsRepository, u TeamUsersRepository, txMgr TransactionManager) *TeamsService {
	return &TeamsService{
		teamsRepo: t,
		usersRepo: u,
		txMgr:     txMgr,
	}
}

func (s *TeamsService) AddTeam(ctx context.Context, team *models.Team) error {
	return s.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.teamsRepo.InsertTeam(txCtx, team.TeamName); err != nil {
			return fmt.Errorf("inserting team: %w", err)
		}

		for _, member := range team.Members {
			if err := s.usersRepo.UpsertUser(txCtx, member.UserID, member.Username, team.TeamName, member.IsActive); err != nil {
				return fmt.Errorf("upserting user %s: %w", member.UserID, err)
			}
		}

		return nil
	})
}

func (s *TeamsService) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	team, err := s.teamsRepo.GetTeam(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("getting team: %w", err)
	}
	return team, nil
}

func (s *TeamsService) GetUserTeam(ctx context.Context, userID string) (string, error) {
	teamName, err := s.teamsRepo.GetUserTeam(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("getting user team: %w", err)
	}
	return teamName, nil
}

func (s *TeamsService) GetActiveTeamMembers(ctx context.Context, teamName string, excludeUserID string) ([]string, error) {
	members, err := s.teamsRepo.GetActiveTeamMembers(ctx, teamName, excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("getting team members: %w", err)
	}
	return members, nil
}
