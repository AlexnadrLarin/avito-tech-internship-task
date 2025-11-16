package models

type User struct {
	UserID   string `json:"user_id" validate:"required,max=255"`
	Username string `json:"username" validate:"required,max=255"`
	TeamName string `json:"team_name" validate:"required,max=255"`
	IsActive bool   `json:"is_active"`
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id" validate:"required,max=255"`
	IsActive bool   `json:"is_active"`
}

type GetUserReviewsQuery struct {
	UserID string `validate:"required,max=255"`
}

type ReviewerStats struct {
	UserID        string `json:"user_id"`
	ReviewsNumber int    `json:"reviews_number"`
}
