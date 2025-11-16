package repository

import (
	"context"
	"fmt"
	"time"

	"pull-request-service/internal/models"
	database "pull-request-service/pkg/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PullRequestRepository struct {
	db *pgxpool.Pool
}

func NewPullRequestRepository(db *pgxpool.Pool) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (repo *PullRequestRepository) CreatePR(ctx context.Context, pr *models.PullRequest) error {
	query := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at) 
		VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now().UTC()
	tx := database.GetTx(ctx, repo.db)
	_, err := tx.Exec(ctx, query, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, now)
	if err != nil {
		return fmt.Errorf("creating pull request: %w", err)
	}

	return nil
}

func (repo *PullRequestRepository) GetPR(ctx context.Context, prID string) (*models.PullRequest, error) {
	query := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at 
		FROM pull_requests 
		WHERE pull_request_id=$1
	`

	var p models.PullRequest
	tx := database.GetTx(ctx, repo.db)

	err := tx.QueryRow(ctx, query, prID).
		Scan(&p.PullRequestID, &p.PullRequestName, &p.AuthorID, &p.Status, &p.CreatedAt, &p.MergedAt)
	if err != nil {
		return nil, fmt.Errorf("getting pull request: %w", err)
	}

	return &p, nil
}

func (repo *PullRequestRepository) MergePR(ctx context.Context, prID string) error {
	query := `
		UPDATE pull_requests 
		SET status='MERGED', merged_at=COALESCE(merged_at,$1) 
		WHERE pull_request_id=$2
	`

	now := time.Now().UTC()
	tx := database.GetTx(ctx, repo.db)

	_, err := tx.Exec(ctx, query, now, prID)
	if err != nil {
		return fmt.Errorf("merging pull request: %w", err)
	}

	return nil
}
