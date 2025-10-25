package account

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/skamranahmed/go-bank/internal/account/types"
	"github.com/skamranahmed/go-bank/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GetAccountsTestSuite struct {
	suite.Suite
	app testutils.TestApp
}

func TestGetAccountsTestSuite(t *testing.T) {
	suite.Run(t, new(GetAccountsTestSuite))
}

// SetupSuite runs once before all tests
func (suite *GetAccountsTestSuite) SetupSuite() {
	suite.app = testutils.NewTestApp(suite.T().Context(), nil, postgresTestContainer, redisTestContainer)

	fixtures, err := testfixtures.New(
		testfixtures.Database(suite.app.Db.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("./fixtures/GetAccounts_test"),
	)
	if err != nil {
		suite.T().Fatal(err)
	}

	err = fixtures.Load()
	if err != nil {
		suite.T().Fatal(err)
	}
}

// TearDownSuite runs once after all tests
func (suite *GetAccountsTestSuite) TearDownSuite() {
	suite.app.TeardownFunc()
}

func (suite *GetAccountsTestSuite) TestMissingAuthorizationHeader() {
	suite.T().Run("missing authorization header returns 401", func(t *testing.T) {
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts", http.MethodGet, nil, nil)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Authorization header is missing")
	})
}

func (suite *GetAccountsTestSuite) TestInvalidAuthorizationHeaderFormat() {
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
			responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts", http.MethodGet, nil, headers)
			assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

			response := testutils.DecodeErrorResponse(t, responseRecorder)
			testutils.AssertFieldError(t, response, "message", tc.expectedErrMessage)
		})
	}
}

func (suite *GetAccountsTestSuite) TestInvalidToken() {
	suite.T().Run("invalid token returns 401", func(t *testing.T) {
		invalidToken := "invalid.jwt.token"
		headers := map[string]string{
			"Authorization": "Bearer " + invalidToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Invalid or expired token")
	})
}

func (suite *GetAccountsTestSuite) TestExpiredToken() {
	suite.T().Run("expired token returns 401", func(t *testing.T) {
		expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcmVzX2F0IjoxNzYxMzMzNTgwLCJpc3N1ZWRfYXQiOjE3NjEzMzI2ODAsInRva2VuX2lkIjoiZmE5ZGI3ZDEtZDQ1MS00NzgzLWE3YTYtYzUxM2E5NDgyNzMwIiwidXNlcl9pZCI6IjhlMTJjYjBhLWQxMjAtNDI0My1iNzRhLWY3NzFhNDZmN2JmZCJ9.wx7guQWcG4cLPEZQcZEQrHrxAC1pNoTFpmwlBk3UUTg"

		headers := map[string]string{
			"Authorization": "Bearer " + expiredToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Invalid or expired token")
	})
}

func (suite *GetAccountsTestSuite) TestSuccessfulGetAccounts() {
	suite.T().Run("valid token returns user accounts", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response types.GetAccountsResponse
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		// verify the user has 2 accounts
		assert.Len(t, response.Data, 2)

		// verify first account (sorted by created_at ASC, so CURRENT_ACCOUNT comes first)
		assert.Equal(t, int64(98765432109876), response.Data[0].ID)
		assert.Equal(t, userID, response.Data[0].UserID)
		assert.Equal(t, int64(98000), response.Data[0].Balance)
		assert.Equal(t, "CURRENT_ACCOUNT", string(response.Data[0].Type))
		assert.NotZero(t, response.Data[0].CreatedAt)
		assert.NotZero(t, response.Data[0].UpdatedAt)

		// verify second account (SAVINGS_ACCOUNT)
		assert.Equal(t, int64(12345678901234), response.Data[1].ID)
		assert.Equal(t, userID, response.Data[1].UserID)
		assert.Equal(t, int64(150000), response.Data[1].Balance)
		assert.Equal(t, "SAVINGS_ACCOUNT", string(response.Data[1].Type))
		assert.NotZero(t, response.Data[1].CreatedAt)
		assert.NotZero(t, response.Data[1].UpdatedAt)
	})
}

func (suite *GetAccountsTestSuite) TestNoAccountsForUser() {
	suite.T().Run("user with no accounts returns empty array", func(t *testing.T) {
		// create a token for a user that doesn't have any accounts
		nonExistentUserID := "00000000-0000-0000-0000-000000000000"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), nonExistentUserID)
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts", http.MethodGet, nil, headers)

		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response types.GetAccountsResponse
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		// verify the response contains an empty array
		assert.NotNil(t, response.Data)
		assert.Len(t, response.Data, 0)
	})
}

func (suite *GetAccountsTestSuite) TestUserCanOnlySeeOwnAccounts() {
	suite.T().Run("user can only see their own accounts", func(t *testing.T) {
		// use the second user who has only 1 account
		userID := "b2c3d4e5-f6a7-5b6c-9d0e-1f2a3b4c5d6e"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response types.GetAccountsResponse
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		// verify the user has only 1 account
		assert.Len(t, response.Data, 1)

		// verify account data
		assert.Equal(t, int64(11111111111111), response.Data[0].ID)
		assert.Equal(t, userID, response.Data[0].UserID)
		assert.Equal(t, int64(50000), response.Data[0].Balance)
		assert.Equal(t, "SAVINGS_ACCOUNT", string(response.Data[0].Type))
	})
}

func (suite *GetAccountsTestSuite) TestResponseFormat() {
	suite.T().Run("response has correct format", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response map[string]interface{}
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		// check that response has "data" field
		data, exists := response["data"]
		assert.True(t, exists, "response should have 'data' field")
		assert.NotNil(t, data)

		dataArray, ok := data.([]interface{})
		assert.True(t, ok, "data should be an array")
		assert.NotEmpty(t, dataArray)

		// verify first account has all required fields
		firstAccount, ok := dataArray[0].(map[string]interface{})
		assert.True(t, ok, "first account should be a map")

		requiredFields := []string{"id", "created_at", "updated_at", "user_id", "balance", "type"}
		for _, field := range requiredFields {
			_, exists := firstAccount[field]
			assert.True(t, exists, "account should contain field: %s", field)
		}
	})
}
