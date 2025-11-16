package repository

import (
	"context"
	"fmt"

	"pull-request-service/internal/models"
	database "pull-request-service/pkg/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(db *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{db: db}
}


func (repo *UsersRepository) UpdateUserActiveStatus(ctx context.Context, userID string, isActive bool) error {
	tx := database.GetTx(ctx, repo.db)

	query := `
		UPDATE users 
		SET is_active=$1 
		WHERE user_id=$2
	`

	_, err := tx.Exec(ctx, query, isActive, userID)
	if err != nil {
		return fmt.Errorf("updating user status: %w", err)
	}

	return nil
}

func (repo *UsersRepository) GetUser(ctx context.Context, userID string) (*models.User, error) {
	var u models.User
	tx := database.GetTx(ctx, repo.db)

	query := `
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE user_id=$1
	`

	err := tx.QueryRow(ctx, query, userID).Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}

	return &u, nil
}

func (repo *UsersRepository) UpsertUser(ctx context.Context, userID, username, teamName string, isActive bool) error {
	tx := database.GetTx(ctx, repo.db)

	query := `
		INSERT INTO users (user_id, username, team_name, is_active) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET 
			username  = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active
	`

	_, err := tx.Exec(ctx, query, userID, username, teamName, isActive)
	if err != nil {
		return fmt.Errorf("upserting user: %w", err)
	}

	return nil
}
