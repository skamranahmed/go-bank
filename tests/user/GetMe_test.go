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

type GetMeTestSuite struct {
	suite.Suite
	app testutils.TestApp
}

func TestGetMeTestSuite(t *testing.T) {
	suite.Run(t, new(GetMeTestSuite))
}

func (suite *GetMeTestSuite) SetupSuite() {
	suite.app = testutils.NewTestApp(suite.T().Context(), nil, postgresTestContainer, redisTestContainer)

	fixtures, err := testfixtures.New(
		testfixtures.Database(suite.app.Db.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("./fixtures/GetMe_test"),
	)
	if err != nil {
		suite.T().Fatal(err)
	}

	err = fixtures.Load()
	if err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *GetMeTestSuite) TearDownSuite() {
	suite.app.TeardownFunc()
}

func (suite *GetMeTestSuite) TestMissingAuthorizationHeader() {
	suite.T().Run("missing authorization header returns 401", func(t *testing.T) {
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodGet, nil, nil)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Authorization header is missing")
	})
}

func (suite *GetMeTestSuite) TestInvalidAuthorizationHeaderFormat() {
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
			headers := map[string]string{
				"Authorization": tc.authHeader,
			}
			responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodGet, nil, headers)
			assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

			response := testutils.DecodeErrorResponse(t, responseRecorder)
			testutils.AssertFieldError(t, response, "message", tc.expectedErrMessage)
		})
	}
}

func (suite *GetMeTestSuite) TestInvalidToken() {
	suite.T().Run("invalid token returns 401", func(t *testing.T) {
		invalidToken := "invalid.jwt.token"
		headers := map[string]string{
			"Authorization": "Bearer " + invalidToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Invalid or expired token")
	})
}

func (suite *GetMeTestSuite) TestExpiredToken() {
	suite.T().Run("expired token returns 401", func(t *testing.T) {
		expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcmVzX2F0IjoxNzYxMzMzNTgwLCJpc3N1ZWRfYXQiOjE3NjEzMzI2ODAsInRva2VuX2lkIjoiZmE5ZGI3ZDEtZDQ1MS00NzgzLWE3YTYtYzUxM2E5NDgyNzMwIiwidXNlcl9pZCI6IjhlMTJjYjBhLWQxMjAtNDI0My1iNzRhLWY3NzFhNDZmN2JmZCJ9.wx7guQWcG4cLPEZQcZEQrHrxAC1pNoTFpmwlBk3UUTg"

		headers := map[string]string{
			"Authorization": "Bearer " + expiredToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Invalid or expired token")
	})
}

func (suite *GetMeTestSuite) TestSuccessfulGetMe() {
	suite.T().Run("valid token returns user information", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response types.GetMeResponse
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, userID, response.Data.ID.String())
		assert.Equal(t, "testuser@example.com", response.Data.Email)
		assert.Equal(t, "test_user_123", response.Data.Username)
		assert.NotZero(t, response.Data.CreatedAt)
		assert.NotZero(t, response.Data.UpdatedAt)
	})
}

func (suite *GetMeTestSuite) TestUserNotFound() {
	suite.T().Run("token with non-existent user ID returns error", func(t *testing.T) {
		// create a token for a user that doesn't exist in the database
		nonExistentUserID := "00000000-0000-0000-0000-000000000000"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), nonExistentUserID)
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodGet, nil, headers)

		assert.Equal(t, http.StatusNotFound, responseRecorder.Code)
	})
}

func (suite *GetMeTestSuite) TestResponseFormat() {
	suite.T().Run("response has correct format", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me", http.MethodGet, nil, headers)
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
