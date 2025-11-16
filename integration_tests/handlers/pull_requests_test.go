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

func TestPullRequestHandler_Integration(t *testing.T) {

	setupTeam := func(teamName string, members []map[string]interface{}) error {
		team := map[string]interface{}{
			"team_name": teamName,
			"members":   members,
		}

		resp, err := helpers.MakeRequest("POST", "/team/add", team)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	}

	t.Run("CreatePR creates new pull request", func(t *testing.T) {
		timestamp := time.Now().UnixNano()
		testID := fmt.Sprintf("pr_create_%s_%d", t.Name(), timestamp)
		teamName := fmt.Sprintf("team_%s", testID)
		authorID := fmt.Sprintf("author_%s", testID)
		reviewerID := fmt.Sprintf("reviewer_%s", testID)

		err := setupTeam(teamName, []map[string]interface{}{
			{"user_id": authorID, "username": authorID, "is_active": true},
			{"user_id": reviewerID, "username": reviewerID, "is_active": true},
		})
		require.NoError(t, err)

		pr := map[string]interface{}{
			"pull_request_id":   testID,
			"pull_request_name": "Test PR",
			"author_id":         authorID,
		}

		resp, err := helpers.MakeRequest("POST", "/pullRequest/create", pr)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = helpers.ParseResponse(resp, &result)
		require.NoError(t, err)
		assert.NotNil(t, result["pr"])
	})

	t.Run("MergePR merges pull request", func(t *testing.T) {
		timestamp := time.Now().UnixNano()
		testID := fmt.Sprintf("pr_merge_%s_%d", t.Name(), timestamp)
		teamName := fmt.Sprintf("team_%s", testID)
		authorID := fmt.Sprintf("author_%s", testID)
		reviewerID := fmt.Sprintf("reviewer_%s", testID)

		err := setupTeam(teamName, []map[string]interface{}{
			{"user_id": authorID, "username": authorID, "is_active": true},
			{"user_id": reviewerID, "username": reviewerID, "is_active": true},
		})
		require.NoError(t, err)

		pr := map[string]interface{}{
			"pull_request_id":   testID,
			"pull_request_name": "Test PR 2",
			"author_id":         authorID,
		}

		resp, err := helpers.MakeRequest("POST", "/pullRequest/create", pr)
		require.NoError(t, err)
		resp.Body.Close()

		mergeReq := map[string]interface{}{
			"pull_request_id": testID,
		}

		resp, err = helpers.MakeRequest("POST", "/pullRequest/merge", mergeReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = helpers.ParseResponse(resp, &result)
		require.NoError(t, err)
		assert.NotNil(t, result["pr"])
	})

	t.Run("ReassignReviewer reassigns reviewer", func(t *testing.T) {
		timestamp := time.Now().UnixNano()
		testID := fmt.Sprintf("pr_reassign_%s_%d", t.Name(), timestamp)
		teamName := fmt.Sprintf("team_%s", testID)
		authorID := fmt.Sprintf("author_%s", testID)
		reviewer1ID := fmt.Sprintf("reviewer1_%s", testID)
		reviewer2ID := fmt.Sprintf("reviewer2_%s", testID)
		reviewer3ID := fmt.Sprintf("reviewer3_%s", testID)

		err := setupTeam(teamName, []map[string]interface{}{
			{"user_id": authorID, "username": authorID, "is_active": true},
			{"user_id": reviewer1ID, "username": reviewer1ID, "is_active": true},
			{"user_id": reviewer2ID, "username": reviewer2ID, "is_active": true},
			{"user_id": reviewer3ID, "username": reviewer3ID, "is_active": true},
		})
		require.NoError(t, err)

		pr := map[string]interface{}{
			"pull_request_id":   testID,
			"pull_request_name": "Test PR 3",
			"author_id":         authorID,
		}

		resp, err := helpers.MakeRequest("POST", "/pullRequest/create", pr)
		require.NoError(t, err)
		defer resp.Body.Close()

		var createResult map[string]interface{}
		err = helpers.ParseResponse(resp, &createResult)
		require.NoError(t, err)

		prData, ok := createResult["pr"].(map[string]interface{})
		require.True(t, ok, "PR data should be present")

		assigned, ok := prData["assigned_reviewers"].([]interface{})
		require.True(t, ok && len(assigned) > 0, "At least one reviewer should be assigned")

		oldReviewerID, ok := assigned[0].(string)
		require.True(t, ok, "Reviewer ID should be a string")

		reassignReq := map[string]interface{}{
			"pull_request_id": testID,
			"old_reviewer_id": oldReviewerID,
		}

		resp, err = helpers.MakeRequest("POST", "/pullRequest/reassign", reassignReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = helpers.ParseResponse(resp, &result)
		require.NoError(t, err)
		assert.NotNil(t, result["pr"])
		assert.NotNil(t, result["replaced_by"])
	})

	t.Run("CreatePR returns 404 for non-existent author", func(t *testing.T) {
		timestamp := time.Now().UnixNano()
		testID := fmt.Sprintf("pr_notfound_%s_%d", t.Name(), timestamp)
		pr := map[string]interface{}{
			"pull_request_id":   testID,
			"pull_request_name": "Test PR 4",
			"author_id":         "nonexistent",
		}

		resp, err := helpers.MakeRequest("POST", "/pullRequest/create", pr)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("MergePR returns 404 for non-existent PR", func(t *testing.T) {
		mergeReq := map[string]interface{}{
			"pull_request_id": "nonexistent",
		}

		resp, err := helpers.MakeRequest("POST", "/pullRequest/merge", mergeReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("CreatePR returns 400 for invalid request", func(t *testing.T) {
		invalidPR := map[string]interface{}{
			"pull_request_id": "",
		}

		resp, err := helpers.MakeRequest("POST", "/pullRequest/create", invalidPR)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
