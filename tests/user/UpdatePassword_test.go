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

type UpdatePasswordTestSuite struct {
	suite.Suite
	app testutils.TestApp
}

func TestUpdatePasswordTestSuite(t *testing.T) {
	suite.Run(t, new(UpdatePasswordTestSuite))
}

func (suite *UpdatePasswordTestSuite) SetupSuite() {
	suite.app = testutils.NewTestApp(suite.T().Context(), nil, postgresTestContainer, redisTestContainer)

	fixtures, err := testfixtures.New(
		testfixtures.Database(suite.app.Db.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("./fixtures/UpdatePassword_test"),
	)
	if err != nil {
		suite.T().Fatal(err)
	}

	err = fixtures.Load()
	if err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *UpdatePasswordTestSuite) TearDownSuite() {
	suite.app.TeardownFunc()
}

func (suite *UpdatePasswordTestSuite) TestMissingAuthorizationHeader() {
	suite.T().Run("missing authorization header returns 401", func(t *testing.T) {
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"current_password": "password",
				"new_password":     "newPassword456",
			},
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me/password", http.MethodPut, payload, nil)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Authorization header is missing")
	})
}

func (suite *UpdatePasswordTestSuite) TestInvalidAuthorizationHeaderFormat() {
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
					"current_password": "password",
					"new_password":     "newPassword456",
				},
			}

			headers := map[string]string{
				"Authorization": tc.authHeader,
			}
			responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me/password", http.MethodPut, payload, headers)
			assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

			response := testutils.DecodeErrorResponse(t, responseRecorder)
			testutils.AssertFieldError(t, response, "message", tc.expectedErrMessage)
		})
	}
}

func (suite *UpdatePasswordTestSuite) TestInvalidToken() {
	suite.T().Run("invalid token returns 401", func(t *testing.T) {
		invalidToken := "invalid.jwt.token"
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"current_password": "password",
				"new_password":     "newPassword456",
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + invalidToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me/password", http.MethodPut, payload, headers)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Invalid or expired token")
	})
}

func (suite *UpdatePasswordTestSuite) TestMissingRequestBody() {
	suite.T().Run("missing request body returns 400", func(t *testing.T) {
		userID := "d4e5f6a7-b8c9-4d5e-8f9a-0b1c2d3e4f5a"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me/password", http.MethodPut, nil, headers)
		assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
	})
}

func (suite *UpdatePasswordTestSuite) TestInvalidRequestBody() {
	tests := []struct {
		name               string
		payload            map[string]interface{}
		field              string
		expectedErrMessage string
	}{
		{
			name: "missing data field",
			payload: map[string]interface{}{
				"current_password": "password123",
				"new_password":     "newPassword456",
			},
			field:              "current_password",
			expectedErrMessage: "current_password is a required field",
		},
		{
			name: "missing current_password field",
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"new_password": "newPassword456",
				},
			},
			field:              "current_password",
			expectedErrMessage: "current_password is a required field",
		},
		{
			name: "empty current_password",
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"current_password": "",
					"new_password":     "newPassword456",
				},
			},
			field:              "current_password",
			expectedErrMessage: "current_password is a required field",
		},
		{
			name: "missing new_password field",
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"current_password": "password",
				},
			},
			field:              "new_password",
			expectedErrMessage: "new_password is a required field",
		},
		{
			name: "empty new_password",
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"current_password": "password",
					"new_password":     "",
				},
			},
			field:              "new_password",
			expectedErrMessage: "new_password is a required field",
		},
		{
			name: "new_password too short",
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"current_password": "password",
					"new_password":     "short",
				},
			},
			field:              "new_password",
			expectedErrMessage: "new_password must be at least 8 characters",
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			userID := "d4e5f6a7-b8c9-4d5e-8f9a-0b1c2d3e4f5a"

			accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
			assert.NoError(t, err)

			headers := map[string]string{
				"Authorization": "Bearer " + accessToken,
			}
			responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me/password", http.MethodPut, tc.payload, headers)
			assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

			response := testutils.DecodeErrorResponse(t, responseRecorder)
			testutils.AssertFieldError(t, response, tc.field, tc.expectedErrMessage)
		})
	}
}

func (suite *UpdatePasswordTestSuite) TestIncorrectCurrentPassword() {
	suite.T().Run("incorrect current password returns 400", func(t *testing.T) {
		userID := "d4e5f6a7-b8c9-4d5e-8f9a-0b1c2d3e4f5a"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"current_password": "wrongPassword",
				"new_password":     "newPassword123",
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me/password", http.MethodPut, payload, headers)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Current password is incorrect")
	})
}

func (suite *UpdatePasswordTestSuite) TestSuccessfulPasswordUpdate() {
	suite.T().Run("valid request successfully updates password", func(t *testing.T) {
		userID := "e5f6a7b8-c9d0-4e5f-9a0b-1c2d3e4f5a6b"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"current_password": "password",
				"new_password":     "newPassword123",
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me/password", http.MethodPut, payload, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response types.UpdatePasswordResponse
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.True(t, response.Success)
	})
}

func (suite *UpdatePasswordTestSuite) TestUserNotFound() {
	suite.T().Run("token with non-existent user ID returns error", func(t *testing.T) {
		// create a token for a user that doesn't exist in the database
		nonExistentUserID := "00000000-0000-0000-0000-000000000000"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), nonExistentUserID)
		assert.NoError(t, err)

		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"current_password": "password",
				"new_password":     "newPassword456",
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me/password", http.MethodPut, payload, headers)

		assert.Equal(t, http.StatusNotFound, responseRecorder.Code)
	})
}

func (suite *UpdatePasswordTestSuite) TestResponseFormat() {
	suite.T().Run("response has correct format", func(t *testing.T) {
		userID := "d4e5f6a7-b8c9-4d5e-8f9a-0b1c2d3e4f5a"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"current_password": "password",
				"new_password":     "newPassword123",
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/me/password", http.MethodPut, payload, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response map[string]interface{}
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		// check that response has "success" field
		success, exists := response["success"]
		assert.True(t, exists, "response should have 'success' field")
		assert.NotNil(t, success)

		// verify success is true
		successBool, ok := success.(bool)
		assert.True(t, ok, "success should be a boolean")
		assert.True(t, successBool, "success should be true")
	})
}
