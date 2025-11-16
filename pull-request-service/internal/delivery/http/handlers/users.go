package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"pull-request-service/internal/delivery/http/helpers"
	"pull-request-service/internal/models"
)

type UsersService interface {
	SetUserActiveStatus(ctx context.Context, userID string, isActive bool) (*models.User, error)
	GetUserReviews(ctx context.Context, userID string) ([]*models.PullRequestShort, error)
	GetReviewsStats(ctx context.Context) ([]*models.ReviewerStats, error)
}

type UsersHandler struct {
	usersService UsersService
	logger       *slog.Logger
	validator    Validator
}

func NewUsersHandler(s UsersService, logger *slog.Logger, validator Validator) *UsersHandler {
	return &UsersHandler{
		usersService: s,
		logger:       logger,
		validator:    validator,
	}
}

func (h *UsersHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req models.SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, models.ErrNotFound, "invalid JSON")
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, models.ErrNotFound, err.Error())
		return
	}

	user, err := h.usersService.SetUserActiveStatus(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		h.logger.Error("set user active status failed", "user_id", req.UserID, "err", err)
		helpers.WriteError(w, http.StatusNotFound, models.ErrNotFound, "user not found")
		return
	}

	helpers.WriteSuccess(w, http.StatusOK, map[string]interface{}{"user": user})
}

func (h *UsersHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	query := models.GetUserReviewsQuery{UserID: userID}
	if err := h.validator.Validate(&query); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, models.ErrNotFound, err.Error())
		return
	}

	prs, err := h.usersService.GetUserReviews(r.Context(), userID)
	if err != nil {
		h.logger.Error("get user reviews failed", "user_id", userID, "err", err)
		helpers.WriteError(w, http.StatusNotFound, models.ErrNotFound, "user not found")
		return
	}

	helpers.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	})
}

func (h *UsersHandler) GetReviewsStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.usersService.GetReviewsStats(r.Context())
	if err != nil {
		h.logger.Error("get reviews stats failed", "err", err)
		helpers.WriteError(w, http.StatusNotFound, models.ErrNoCandidate, "no found users")
		return
	}

	helpers.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"reviews_stats_list": stats,
	})
}
