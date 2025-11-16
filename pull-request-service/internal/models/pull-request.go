package models

import "time"

type PullRequestStatus string

const (
	StatusOpen   PullRequestStatus = "OPEN"
	StatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID   string            `json:"pull_request_id" validate:"required,max=255"`
	PullRequestName string            `json:"pull_request_name" validate:"required,max=255"`
	AuthorID        string            `json:"author_id" validate:"required,max=255"`
	Status          PullRequestStatus `json:"status" validate:"required,status_enum"`
	Assigned        []string          `json:"assigned_reviewers" validate:"required,max=2"`
	CreatedAt       *time.Time        `json:"createdAt,omitempty"`
	MergedAt        *time.Time        `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	PullRequestID   string            `json:"pull_request_id" validate:"required,max=255"`
	PullRequestName string            `json:"pull_request_name" validate:"required,max=255"`
	AuthorID        string            `json:"author_id" validate:"required,max=255"`
	Status          PullRequestStatus `json:"status" validate:"required,status_enum"`
}

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id" validate:"required,max=255"`
	PullRequestName string `json:"pull_request_name" validate:"required,max=255"`
	AuthorID        string `json:"author_id" validate:"required,max=255"`
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required,max=255"`
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required,max=255"`
	OldReviewerID string `json:"old_reviewer_id" validate:"required,max=255"`
}
