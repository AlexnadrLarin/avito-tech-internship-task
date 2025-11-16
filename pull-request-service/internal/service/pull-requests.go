package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"pull-request-service/internal/models"
)

type PullRequestRepository interface {
	CreatePR(ctx context.Context, pr *models.PullRequest) error
	GetPR(ctx context.Context, prID string) (*models.PullRequest, error)
	MergePR(ctx context.Context, prID string) error
}

type ReviewRepository interface {
	AddReviewer(ctx context.Context, prID, userID string) error
	RemoveReviewer(ctx context.Context, prID, userID string) error
	GetPRReviewers(ctx context.Context, prID string) ([]string, error)
	GetPRsByReviewer(ctx context.Context, userID string) ([]*models.PullRequestShort, error)
}

type TeamInfoRepository interface {
	GetUserTeam(ctx context.Context, userID string) (string, error)
	GetActiveTeamMembers(ctx context.Context, teamName string, excludeUserID string) ([]string, error)
}

type PullRequestService struct {
	prRepo     PullRequestRepository
	reviewRepo ReviewRepository
	teamsRepo  TeamInfoRepository
	txMgr      TransactionManager
}

func NewPullRequestService(
	prRepo PullRequestRepository,
	reviewRepo ReviewRepository,
	teamsRepo TeamInfoRepository,
	txMgr TransactionManager,
) *PullRequestService {
	return &PullRequestService{
		prRepo:     prRepo,
		reviewRepo: reviewRepo,
		teamsRepo:  teamsRepo,
		txMgr:      txMgr,
	}
}

func pickRandom(src []string, n int) []string {
	if len(src) == 0 || n <= 0 {
		return nil
	}
	if len(src) <= n {
		return append([]string{}, src...)
	}

	out := append([]string{}, src...)
	for i := range out {
		j := i + int(time.Now().UnixNano()%int64(len(out)-i))
		out[i], out[j] = out[j], out[i]
	}
	return out[:n]
}

func (s *PullRequestService) CreatePR(ctx context.Context, pr *models.PullRequest) (*models.PullRequest, error) {
	if pr.Status == "" {
		pr.Status = models.StatusOpen
	}

	var result *models.PullRequest

	err := s.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.prRepo.CreatePR(txCtx, pr); err != nil {
			return fmt.Errorf("creating PR: %w", err)
		}

		teamName, err := s.teamsRepo.GetUserTeam(txCtx, pr.AuthorID)
		if err != nil {
			return fmt.Errorf("getting author team: %w", err)
		}

		candidates, err := s.teamsRepo.GetActiveTeamMembers(txCtx, teamName, pr.AuthorID)
		if err != nil {
			return fmt.Errorf("getting team members: %w", err)
		}

		assigned := pickRandom(candidates, 2)
		for _, reviewerID := range assigned {
			if err := s.reviewRepo.AddReviewer(txCtx, pr.PullRequestID, reviewerID); err != nil {
				return fmt.Errorf("adding reviewer: %w", err)
			}
		}

		result, err = s.getPRWithReviewers(txCtx, pr.PullRequestID)
		if err != nil {
			return fmt.Errorf("getting created PR: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *PullRequestService) getPRWithReviewers(ctx context.Context, prID string) (*models.PullRequest, error) {
	pr, err := s.prRepo.GetPR(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("getting PR: %w", err)
	}

	reviewers, err := s.reviewRepo.GetPRReviewers(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("getting reviewers: %w", err)
	}
	pr.Assigned = reviewers

	return pr, nil
}

func (s *PullRequestService) GetPR(ctx context.Context, prID string) (*models.PullRequest, error) {
	return s.getPRWithReviewers(ctx, prID)
}

func (s *PullRequestService) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	var result *models.PullRequest

	err := s.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.prRepo.MergePR(txCtx, prID); err != nil {
			return fmt.Errorf("merging PR: %w", err)
		}

		var err error
		result, err = s.getPRWithReviewers(txCtx, prID)
		if err != nil {
			return fmt.Errorf("getting merged PR: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *PullRequestService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*models.PullRequest, string, error) {
	var result *models.PullRequest
	var newReviewerID string

	err := s.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		pr, err := s.getPRWithReviewers(txCtx, prID)
		if err != nil {
			return fmt.Errorf("getting PR: %w", err)
		}

		if pr.Status == models.StatusMerged {
			return errors.New("PR_MERGED")
		}

		if !slices.Contains(pr.Assigned, oldReviewerID) {
			return errors.New("NOT_ASSIGNED")
		}

		teamName, err := s.teamsRepo.GetUserTeam(txCtx, oldReviewerID)
		if err != nil {
			return fmt.Errorf("getting reviewer team: %w", err)
		}

		candidates, err := s.teamsRepo.GetActiveTeamMembers(txCtx, teamName, oldReviewerID)
		if err != nil {
			return fmt.Errorf("getting team members: %w", err)
		}

		filteredCandidates := make([]string, 0)
		for _, candidate := range candidates {
			if slices.Contains(pr.Assigned, candidate) {
				continue
			}

			if candidate == pr.AuthorID {
				continue
			}
			filteredCandidates = append(filteredCandidates, candidate)
		}

		if len(filteredCandidates) == 0 {
			return errors.New("NO_CANDIDATE")
		}

		selected := pickRandom(filteredCandidates, 1)
		if len(selected) == 0 {
			return errors.New("NO_CANDIDATE")
		}
		newReviewerID = selected[0]

		if err := s.reviewRepo.RemoveReviewer(txCtx, prID, oldReviewerID); err != nil {
			return fmt.Errorf("removing old reviewer: %w", err)
		}

		if err := s.reviewRepo.AddReviewer(txCtx, prID, newReviewerID); err != nil {
			return fmt.Errorf("adding new reviewer: %w", err)
		}

		result, err = s.getPRWithReviewers(txCtx, prID)
		if err != nil {
			return fmt.Errorf("getting updated PR: %w", err)
		}

		return nil
	})

	if err != nil {
		switch err.Error() {
		case "PR_MERGED":
			return nil, "", fmt.Errorf("error: code: PR_MERGED, message: cannot reassign on merged PR")
		case "NOT_ASSIGNED":
			return nil, "", fmt.Errorf("error: code: NOT_ASSIGNED, message: reviewer is not assigned to this PR")
		case "NO_CANDIDATE":
			return nil, "", fmt.Errorf("error: code: NO_CANDIDATE, message: no active replacement candidate in team")
		default:
			return nil, "", err
		}
	}

	return result, newReviewerID, nil
}
