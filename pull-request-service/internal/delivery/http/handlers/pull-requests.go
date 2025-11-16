package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"pull-request-service/internal/delivery/http/helpers"
	"pull-request-service/internal/models"
)

type PullRequestService interface {
	CreatePR(ctx context.Context, pr *models.PullRequest) (*models.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*models.PullRequest, string, error)
}

type PullRequestHandler struct {
	prService PullRequestService
	logger    *slog.Logger
	validator Validator
}

func NewPullRequestHandler(s PullRequestService, logger *slog.Logger, validator Validator) *PullRequestHandler {
	return &PullRequestHandler{
		prService: s,
		logger:    logger,
		validator: validator,
	}
}

func (h *PullRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePRRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, models.ErrNotFound, "invalid JSON")
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, models.ErrNotFound, err.Error())
		return
	}

	pr := &models.PullRequest{
		PullRequestID:   req.PullRequestID,
		PullRequestName: req.PullRequestName,
		AuthorID:        req.AuthorID,
		Status:          models.StatusOpen,
	}

	result, err := h.prService.CreatePR(r.Context(), pr)
	if err != nil {
		if strings.Contains(err.Error(), "conflict") || strings.Contains(err.Error(), "duplicate") {
			helpers.WriteError(w, http.StatusConflict, models.ErrPRExists, "PR id already exists")
			return
		}
		h.logger.Error("create PR failed", "pr_id", req.PullRequestID, "err", err)
		helpers.WriteError(w, http.StatusNotFound, models.ErrNotFound, "author/team not found")
		return
	}

	helpers.WriteSuccess(w, http.StatusCreated, map[string]interface{}{"pr": result})
}

func (h *PullRequestHandler) Merge(w http.ResponseWriter, r *http.Request) {
	var req models.MergePRRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, models.ErrNotFound, "invalid JSON")
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, models.ErrNotFound, err.Error())
		return
	}

	result, err := h.prService.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		h.logger.Error("merge PR failed", "pr_id", req.PullRequestID, "err", err)
		helpers.WriteError(w, http.StatusNotFound, models.ErrNotFound, "PR not found")
		return
	}

	helpers.WriteSuccess(w, http.StatusOK, map[string]interface{}{"pr": result})
}

func (h *PullRequestHandler) Reassign(w http.ResponseWriter, r *http.Request) {
	var req models.ReassignReviewerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, models.ErrNotFound, "invalid JSON")
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, models.ErrNotFound, err.Error())
		return
	}

	result, newReviewerID, err := h.prService.ReassignReviewer(r.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "PR_MERGED") {
			helpers.WriteError(w, http.StatusConflict, models.ErrPRMerged, "cannot reassign on merged PR")
			return
		}
		if strings.Contains(errStr, "NOT_ASSIGNED") {
			helpers.WriteError(w, http.StatusConflict, models.ErrNotAssigned, "reviewer is not assigned to this PR")
			return
		}
		if strings.Contains(errStr, "NO_CANDIDATE") {
			helpers.WriteError(w, http.StatusConflict, models.ErrNoCandidate, "no active replacement candidate in team")
			return
		}
		h.logger.Error("reassign reviewer failed", "pr_id", req.PullRequestID, "err", err)
		helpers.WriteError(w, http.StatusNotFound, models.ErrNotFound, "PR or user not found")
		return
	}

	helpers.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"pr":          result,
		"replaced_by": newReviewerID,
	})
}
