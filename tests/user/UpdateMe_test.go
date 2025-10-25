package user

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/skamranahmed/go-bank/internal/user/types"
	"github.com/skamranahmed/go-bank/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UpdateMeTestSuite struct {
	suite.Suite
	app testutils.TestApp
}

func TestUpdateMeTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateMeTestSuite))
}

func (suite *UpdateMeTestSuite) SetupSuite() {
	suite.app = testutils.NewTestApp(suite.T().Context(), nil, postgresTestContainer, redisTestContainer)

	fixtures, err := testfixtures.New(
		testfixtures.Database(suite.app.Db.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("./fixtures/UpdateMe_test"),
	)
	if err != nil {
		suite.T().Fatal(err)
	}

	err = fixtures.Load()
	if err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *UpdateMeTestSuite) TearDownSuite() {
	suite.app.TeardownFunc()
}

func (suite *UpdateMeTestSuite) TestMissingAuthorizationHeader() {
	suite.T().Run("missing authorization header returns 401", func(t *testing.T) {
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"username": "new_username",
			},
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodPatch, payload, nil)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Authorization header is missing")
	})
}

func (suite *UpdateMeTestSuite) TestInvalidAuthorizationHeaderFormat() {
	tests := []struct {
		name               string
		authHeader         string
		expectedErrMessage string
	}{
		{
			name:               "authorization header without Bearer prefix",
			authHeader:         "some-token",
			expectedErrMessage: "Invalid Authorization header format",
		},
		{
			name:               "authorization header with invalid format",
			authHeader:         "InvalidFormat token",
			expectedErrMessage: "Invalid Authorization header format",
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"data": map[string]interface{}{
					"username": "new_username",
				},
			}

			headers := map[string]string{
				"Authorization": tc.authHeader,
			}
			responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodPatch, payload, headers)
			assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

			response := testutils.DecodeErrorResponse(t, responseRecorder)
			testutils.AssertFieldError(t, response, "message", tc.expectedErrMessage)
		})
	}
}

func (suite *UpdateMeTestSuite) TestInvalidToken() {
	suite.T().Run("invalid token returns 401", func(t *testing.T) {
		invalidToken := "invalid.jwt.token"
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"username": "new_username",
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + invalidToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodPatch, payload, headers)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Invalid or expired token")
	})
}

func (suite *UpdateMeTestSuite) TestExpiredToken() {
	suite.T().Run("expired token returns 401", func(t *testing.T) {
		expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcmVzX2F0IjoxNzYxMzMzNTgwLCJpc3N1ZWRfYXQiOjE3NjEzMzI2ODAsInRva2VuX2lkIjoiZmE5ZGI3ZDEtZDQ1MS00NzgzLWE3YTYtYzUxM2E5NDgyNzMwIiwidXNlcl9pZCI6IjhlMTJjYjBhLWQxMjAtNDI0My1iNzRhLWY3NzFhNDZmN2JmZCJ9.wx7guQWcG4cLPEZQcZEQrHrxAC1pNoTFpmwlBk3UUTg"

		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"username": "new_username",
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + expiredToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodPatch, payload, headers)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Invalid or expired token")
	})
}

func (suite *UpdateMeTestSuite) TestMissingRequestBody() {
	suite.T().Run("missing request body returns 400", func(t *testing.T) {
		userID := "b2c3d4e5-f6a7-4b5c-8d9e-0f1a2b3c4d5e"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodPatch, nil, headers)
		assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	})
}

func (suite *UpdateMeTestSuite) TestInvalidRequestBody() {
	tests := []struct {
		name               string
		payload            map[string]interface{}
		field              string
		expectedErrMessage string
	}{
		{
			name: "missing data field",
			payload: map[string]interface{}{
				"username": "new_username",
			},
			field:              "username",
			expectedErrMessage: "username is a required field",
		},
		{
			name: "missing username field",
			payload: map[string]interface{}{
				"data": map[string]interface{}{},
			},
			field:              "username",
			expectedErrMessage: "username is a required field",
		},
		{
			name: "empty username",
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"username": "",
				},
			},
			field:              "username",
			expectedErrMessage: "username is a required field",
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			userID := "b2c3d4e5-f6a7-4b5c-8d9e-0f1a2b3c4d5e"

			accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
			assert.NoError(t, err)

			headers := map[string]string{
				"Authorization": "Bearer " + accessToken,
			}
			responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodPatch, tc.payload, headers)
			assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

			response := testutils.DecodeErrorResponse(t, responseRecorder)
			testutils.AssertFieldError(t, response, tc.field, tc.expectedErrMessage)
		})
	}
}

func (suite *UpdateMeTestSuite) TestSuccessfulUpdateUser() {
	suite.T().Run("valid request successfully updates username", func(t *testing.T) {
		userID := "b2c3d4e5-f6a7-4b5c-8d9e-0f1a2b3c4d5e"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"username": "updated_username_123",
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodPatch, payload, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response types.UpdateUserResponse
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, userID, response.Data.ID.String())
		assert.Equal(t, "updated_username_123", response.Data.Username)
		assert.Equal(t, "updateuser@example.com", response.Data.Email)
		assert.NotZero(t, response.Data.CreatedAt)
		assert.NotZero(t, response.Data.UpdatedAt)
	})
}

func (suite *UpdateMeTestSuite) TestUserNotFound() {
	suite.T().Run("token with non-existent user ID returns error", func(t *testing.T) {
		// create a token for a user that doesn't exist in the database
		nonExistentUserID := "00000000-0000-0000-0000-000000000000"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), nonExistentUserID)
		assert.NoError(t, err)

		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"username": "new_username",
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodPatch, payload, headers)

		assert.Equal(t, http.StatusNotFound, responseRecorder.Code)
	})
}

func (suite *UpdateMeTestSuite) TestDuplicateUsername() {
	suite.T().Run("updating to existing username returns conflict error", func(t *testing.T) {
		userID := "b2c3d4e5-f6a7-4b5c-8d9e-0f1a2b3c4d5e"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		// try to update to a username that already exists (from another user in fixtures)
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"username": "existing_user",
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodPatch, payload, headers)
		assert.Equal(t, http.StatusConflict, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "This username is already in use. Please choose another.")
	})
}

func (suite *UpdateMeTestSuite) TestResponseFormat() {
	suite.T().Run("response has correct format", func(t *testing.T) {
		userID := "b2c3d4e5-f6a7-4b5c-8d9e-0f1a2b3c4d5e"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"username": "format_test_user",
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodPatch, payload, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response map[string]interface{}
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		// check that response has "data" field
		data, exists := response["data"]
		assert.True(t, exists, "response should have 'data' field")
		assert.NotNil(t, data)

		dataMap, ok := data.(map[string]interface{})
		assert.True(t, ok, "data should be a map")

		// verify all required fields exist
		requiredFields := []string{"id", "created_at", "updated_at", "email", "username"}
		for _, field := range requiredFields {
			_, exists := dataMap[field]
			assert.True(t, exists, "data should contain field: %s", field)
		}

		// verify password is NOT included in response
		_, passwordExists := dataMap["password"]
		assert.False(t, passwordExists, "data should NOT contain password field")
	})
}
