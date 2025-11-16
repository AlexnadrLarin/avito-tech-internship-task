package repository

import (
	"context"
	"fmt"

	"pull-request-service/internal/models"
	database "pull-request-service/pkg/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewRepository struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (repo *ReviewRepository) AddReviewer(ctx context.Context, prID, userID string) error {
	query := `
		INSERT INTO pr_reviewers (pull_request_id, user_id) 
		VALUES ($1, $2)
	`

	tx := database.GetTx(ctx, repo.db)
	_, err := tx.Exec(ctx, query, prID, userID)
	if err != nil {
		return fmt.Errorf("adding reviewer: %w", err)
	}

	return nil
}

func (repo *ReviewRepository) RemoveReviewer(ctx context.Context, prID, userID string) error {
	query := `
		DELETE FROM pr_reviewers 
		WHERE pull_request_id=$1 AND user_id=$2
	`

	tx := database.GetTx(ctx, repo.db)
	_, err := tx.Exec(ctx, query, prID, userID)
	if err != nil {
		return fmt.Errorf("removing reviewer: %w", err)
	}

	return nil
}

func (repo *ReviewRepository) GetPRReviewers(ctx context.Context, prID string) ([]string, error) {
	query := `
		SELECT user_id 
		FROM pr_reviewers 
		WHERE pull_request_id=$1
	`

	tx := database.GetTx(ctx, repo.db)
	rows, err := tx.Query(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("querying reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scanning reviewer: %w", err)
		}
		reviewers = append(reviewers, uid)
	}

	return reviewers, nil
}

func (repo *ReviewRepository) GetPRsByReviewer(ctx context.Context, userID string) ([]*models.PullRequestShort, error) {
	tx := database.GetTx(ctx, repo.db)

	query := `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := tx.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying PRs by reviewer: %w", err)
	}
	defer rows.Close()

	var prs []*models.PullRequestShort
	for rows.Next() {
		var pr models.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("scanning PR: %w", err)
		}
		prs = append(prs, &pr)
	}

	return prs, nil
}

func (repo *ReviewRepository) GetReviewsStats(ctx context.Context) ([]*models.ReviewerStats, error) {
	tx := database.GetTx(ctx, repo.db)

	query := `
		SELECT user_id, COUNT(*)
		FROM pr_reviewers 
		GROUP BY user_id
	`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying reviews stats: %w", err)
	}
	defer rows.Close()

	var rsl []*models.ReviewerStats
	for rows.Next() {
		var rs models.ReviewerStats
		if err := rows.Scan(&rs.UserID, &rs.ReviewsNumber); err != nil {
			return nil, fmt.Errorf("scanning user reviews stats: %w", err)
		}
		rsl = append(rsl, &rs)
	}

	return rsl, nil
}
