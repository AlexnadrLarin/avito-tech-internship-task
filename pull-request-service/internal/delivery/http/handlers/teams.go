package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"pull-request-service/internal/delivery/http/helpers"
	"pull-request-service/internal/models"
)

type TeamService interface {
	AddTeam(ctx context.Context, team *models.Team) error
	GetTeam(ctx context.Context, teamName string) (*models.Team, error)
}

type TeamHandler struct {
	teamService TeamService
	logger      *slog.Logger
	validator   Validator
}

func NewTeamHandler(s TeamService, logger *slog.Logger, validator Validator) *TeamHandler {
	return &TeamHandler{
		teamService: s,
		logger:      logger,
		validator:   validator,
	}
}

func (h *TeamHandler) Add(w http.ResponseWriter, r *http.Request) {
	var req models.Team
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, models.ErrNotFound, "invalid JSON")
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, models.ErrNotFound, err.Error())
		return
	}

	err := h.teamService.AddTeam(r.Context(), &req)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, models.ErrTeamExists, "team_name already exists")
		return
	}

	team, err := h.teamService.GetTeam(r.Context(), req.TeamName)
	if err != nil {
		h.logger.Error("failed to get created team", "team", req.TeamName, "err", err)
		helpers.WriteError(w, http.StatusInternalServerError, models.ErrNotFound, "failed to get created team")
		return
	}

	helpers.WriteSuccess(w, http.StatusCreated, map[string]interface{}{"team": team})
}

func (h *TeamHandler) Get(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")

	query := models.GetTeamQuery{TeamName: teamName}
	if err := h.validator.Validate(&query); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, models.ErrNotFound, err.Error())
		return
	}

	resp, err := h.teamService.GetTeam(r.Context(), teamName)
	if err != nil {
		helpers.WriteError(w, http.StatusNotFound, models.ErrNotFound, "team not found")
		return
	}

	helpers.WriteSuccess(w, http.StatusOK, resp)
}
