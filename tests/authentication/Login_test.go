package authentication

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/skamranahmed/go-bank/internal/authentication/dto"
	"github.com/skamranahmed/go-bank/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LoginTestSuite struct {
	suite.Suite
	app testutils.TestApp
}

func TestLoginTestSuite(t *testing.T) {
	suite.Run(t, new(LoginTestSuite))
}

func (suite *LoginTestSuite) SetupSuite() {
	suite.app = testutils.NewTestApp(suite.T().Context(), nil, postgresTestContainer, redisTestContainer)

	fixtures, err := testfixtures.New(
		testfixtures.Database(suite.app.Db.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("./fixtures/Login_test"),
	)
	if err != nil {
		suite.T().Fatal(err)
	}

	err = fixtures.Load()
	if err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *LoginTestSuite) TearDownSuite() {
	suite.app.TeardownFunc()
}

func (suite *LoginTestSuite) TestValidationErrors() {
	tests := []struct {
		name               string
		payload            dto.LoginRequest
		field              string
		errMessage         string
		expectedStatusCode int
	}{
		{
			name: "missing username",
			payload: dto.LoginRequest{
				Data: dto.LoginData{
					Password: "password",
				},
			},
			field:              "username",
			errMessage:         "username is a required field",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "missing password",
			payload: dto.LoginRequest{
				Data: dto.LoginData{
					Username: "username",
				},
			},
			field:              "password",
			errMessage:         "password is a required field",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/login", http.MethodPost, tc.payload)
			assert.Equal(t, tc.expectedStatusCode, responseRecorder.Code)

			response := testutils.DecodeErrorResponse(t, responseRecorder)
			testutils.AssertFieldError(t, response, tc.field, tc.errMessage)
		})
	}
}

func (suite *LoginTestSuite) TestAuthenticationErrors() {
	tests := []struct {
		name               string
		payload            dto.LoginRequest
		errMessage         string
		expectedStatusCode int
	}{
		{
			name: "invalid username",
			payload: dto.LoginRequest{
				Data: dto.LoginData{
					Username: "nonexistent_user",
					Password: "password",
				},
			},
			errMessage:         "Invalid username or password.",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "invalid password",
			payload: dto.LoginRequest{
				Data: dto.LoginData{
					Username: "kamran_ahmed",
					Password: "wrongpassword",
				},
			},
			errMessage:         "Invalid username or password.",
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/login", http.MethodPost, tc.payload)
			assert.Equal(t, tc.expectedStatusCode, responseRecorder.Code)

			response := testutils.DecodeErrorResponse(t, responseRecorder)
			testutils.AssertFieldError(t, response, "message", tc.errMessage)
		})
	}
}

func (suite *LoginTestSuite) TestSuccessfulLogin() {
	type SuccessResponse struct {
		AccessToken string `json:"access_token"`
	}

	suite.T().Run("successful login returns access token", func(t *testing.T) {
		userID := "b35ac310-9fa2-40e1-be39-553b07d6235a"

		payload := dto.LoginRequest{
			Data: dto.LoginData{
				Username: "kamran_ahmed",
				Password: "password",
			},
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/login", http.MethodPost, payload)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response SuccessResponse
		err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.AccessToken)

		// verify the access token that is returned in response and check its existence in cache
		tokenData, err := suite.app.Services.AuthenticationService.VerifyAccessToken(t.Context(), response.AccessToken)
		assert.Equal(t, nil, err)
		assert.NotZero(t, tokenData.UserID)
		assert.NotZero(t, tokenData.TokenID)

		assert.Equal(t, userID, tokenData.UserID)

		accessTokenCacheKey := fmt.Sprintf("auth:access_token_id:%s:user_id:%s", tokenData.TokenID, userID)
		tokenInCache, err := suite.app.Cache.Get(t.Context(), accessTokenCacheKey)
		assert.Equal(t, nil, err)
		assert.Equal(t, "", tokenInCache)
	})
}
