package repository

import (
	"context"
	"fmt"

	"pull-request-service/internal/models"
	database "pull-request-service/pkg/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamsRepository struct {
	db *pgxpool.Pool
}

func NewTeamsRepository(db *pgxpool.Pool) *TeamsRepository {
	return &TeamsRepository{db: db}
}

func (repo *TeamsRepository) InsertTeam(ctx context.Context, teamName string) error {
	tx := database.GetTx(ctx, repo.db)

	query := `
		INSERT INTO teams (team_name) 
		VALUES ($1) 
		ON CONFLICT (team_name) DO NOTHING
	`

	_, err := tx.Exec(ctx, query, teamName)
	if err != nil {
		return fmt.Errorf("inserting team: %w", err)
	}

	return nil
}

func (repo *TeamsRepository) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	t := models.Team{}
	tx := database.GetTx(ctx, repo.db)

	teamQuery := `
		SELECT team_name
		FROM teams
		WHERE team_name=$1
	`

	err := tx.QueryRow(ctx, teamQuery, teamName).Scan(&t.TeamName)
	if err != nil {
		return nil, fmt.Errorf("selecting team: %w", err)
	}


	membersQuery := `
		SELECT user_id, username, is_active 
		FROM users 
		WHERE team_name=$1
	`

	rows, err := tx.Query(ctx, membersQuery, teamName)
	if err != nil {
		return nil, fmt.Errorf("selecting team members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var u models.TeamMember

		err := rows.Scan(&u.UserID, &u.Username, &u.IsActive)
		if err != nil {
			return nil, fmt.Errorf("scanning team member: %w", err)
		}
		t.Members = append(t.Members, u)
	}

	return &t, nil
}

func (repo *TeamsRepository) GetUserTeam(ctx context.Context, userID string) (string, error) {
	query := `
		SELECT team_name 
		FROM users 
		WHERE user_id=$1
	`

	var teamName string
	tx := database.GetTx(ctx, repo.db)
	err := tx.QueryRow(ctx, query, userID).Scan(&teamName)
	if err != nil {
		return "", fmt.Errorf("getting user team: %w", err)
	}

	return teamName, nil
}

func (repo *TeamsRepository) GetActiveTeamMembers(ctx context.Context, teamName string, excludeUserID string) ([]string, error) {
	query := `
		SELECT user_id 
		FROM users 
		WHERE team_name=$1 AND is_active=true AND user_id<>$2
	`

	tx := database.GetTx(ctx, repo.db)
	rows, err := tx.Query(ctx, query, teamName, excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("querying team members: %w", err)
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scanning team member: %w", err)
		}
		members = append(members, uid)
	}

	return members, nil
}
