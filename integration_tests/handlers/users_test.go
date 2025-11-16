package handlers_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"integration-tests/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersHandler_Integration(t *testing.T) {

	t.Run("SetIsActive updates user status", func(t *testing.T) {
		timestamp := time.Now().UnixNano()
		testID := fmt.Sprintf("user_setactive_%s_%d", t.Name(), timestamp)
		teamName := fmt.Sprintf("team_%s", testID)
		userID := fmt.Sprintf("user_%s", testID)

		team := map[string]interface{}{
			"team_name": teamName,
			"members": []map[string]interface{}{
				{
					"user_id":   userID,
					"username":  userID,
					"is_active": true,
				},
			},
		}

		resp, err := helpers.MakeRequest("POST", "/team/add", team)
		require.NoError(t, err)
		resp.Body.Close()

		setActiveReq := map[string]interface{}{
			"user_id":   userID,
			"is_active": false,
		}

		resp, err = helpers.MakeRequest("POST", "/users/setIsActive", setActiveReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = helpers.ParseResponse(resp, &result)
		require.NoError(t, err)
		assert.NotNil(t, result["user"])
	})

	t.Run("GetReview returns user reviews", func(t *testing.T) {
		timestamp := time.Now().UnixNano()
		testID := fmt.Sprintf("user_review_%s_%d", t.Name(), timestamp)
		teamName := fmt.Sprintf("team_%s", testID)
		authorID := fmt.Sprintf("author_%s", testID)
		reviewerID := fmt.Sprintf("reviewer_%s", testID)
		prID := fmt.Sprintf("pr_%s", testID)

		team := map[string]interface{}{
			"team_name": teamName,
			"members": []map[string]interface{}{
				{
					"user_id":   authorID,
					"username":  authorID,
					"is_active": true,
				},
				{
					"user_id":   reviewerID,
					"username":  reviewerID,
					"is_active": true,
				},
			},
		}

		resp, err := helpers.MakeRequest("POST", "/team/add", team)
		require.NoError(t, err)
		resp.Body.Close()

		pr := map[string]interface{}{
			"pull_request_id":   prID,
			"pull_request_name": "Test PR",
			"author_id":         authorID,
		}

		resp, err = helpers.MakeRequest("POST", "/pullRequest/create", pr)
		require.NoError(t, err)
		resp.Body.Close()

		resp, err = helpers.MakeRequest("GET", fmt.Sprintf("/users/getReview?user_id=%s", reviewerID), nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = helpers.ParseResponse(resp, &result)
		require.NoError(t, err)
		assert.NotNil(t, result["pull_requests"])
	})

	t.Run("GetReviewsStats returns statistics", func(t *testing.T) {
		resp, err := helpers.MakeRequest("GET", "/users/getReviewsStats", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = helpers.ParseResponse(resp, &result)
		require.NoError(t, err)
		assert.NotNil(t, result["reviews_stats_list"])
	})

	t.Run("SetIsActive returns 404 for non-existent user", func(t *testing.T) {
		setActiveReq := map[string]interface{}{
			"user_id":   "nonexistent",
			"is_active": false,
		}

		resp, err := helpers.MakeRequest("POST", "/users/setIsActive", setActiveReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
