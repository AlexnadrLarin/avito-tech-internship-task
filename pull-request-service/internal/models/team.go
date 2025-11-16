package models

type TeamMember struct {
	UserID   string `json:"user_id" validate:"required,max=255"`
	Username string `json:"username" validate:"required,max=255"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	TeamName string       `json:"team_name" validate:"required,max=255"`
	Members  []TeamMember `json:"members" validate:"required,dive"`
}

type GetTeamQuery struct {
	TeamName string `validate:"required,max=255"`
}
