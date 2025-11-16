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

func TestTeamHandler_Integration(t *testing.T) {

	t.Run("AddTeam creates new team", func(t *testing.T) {
		timestamp := time.Now().UnixNano()
		testID := fmt.Sprintf("team_add_%s_%d", t.Name(), timestamp)
		teamName := fmt.Sprintf("team_%s", testID)
		user1ID := fmt.Sprintf("user1_%s", testID)
		user2ID := fmt.Sprintf("user2_%s", testID)

		team := map[string]interface{}{
			"team_name": teamName,
			"members": []map[string]interface{}{
				{
					"user_id":   user1ID,
					"username":  user1ID,
					"is_active": true,
				},
				{
					"user_id":   user2ID,
					"username":  user2ID,
					"is_active": true,
				},
			},
		}

		resp, err := helpers.MakeRequest("POST", "/team/add", team)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = helpers.ParseResponse(resp, &result)
		require.NoError(t, err)
		assert.NotNil(t, result["team"])
	})

	t.Run("GetTeam returns team", func(t *testing.T) {
		timestamp := time.Now().UnixNano()
		testID := fmt.Sprintf("team_get_%s_%d", t.Name(), timestamp)
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

		resp, err = helpers.MakeRequest("GET", fmt.Sprintf("/team/get?team_name=%s", teamName), nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = helpers.ParseResponse(resp, &result)
		require.NoError(t, err)
		assert.NotNil(t, result["team_name"])
	})

	t.Run("GetTeam returns 404 for non-existent team", func(t *testing.T) {
		resp, err := helpers.MakeRequest("GET", "/team/get?team_name=nonexistent", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("AddTeam returns error for invalid request", func(t *testing.T) {
		invalidTeam := map[string]interface{}{
			"team_name": "",
		}

		resp, err := helpers.MakeRequest("POST", "/team/add", invalidTeam)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
