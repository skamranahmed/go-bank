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

type GetAccountByIDTestSuite struct {
	suite.Suite
	app testutils.TestApp
}

func TestGetAccountByIDTestSuite(t *testing.T) {
	suite.Run(t, new(GetAccountByIDTestSuite))
}

// SetupSuite runs once before all tests
func (suite *GetAccountByIDTestSuite) SetupSuite() {
	suite.app = testutils.NewTestApp(suite.T().Context(), nil, postgresTestContainer, redisTestContainer)

	fixtures, err := testfixtures.New(
		testfixtures.Database(suite.app.Db.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("./fixtures/GetAccountByID_test"),
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
func (suite *GetAccountByIDTestSuite) TearDownSuite() {
	suite.app.TeardownFunc()
}

func (suite *GetAccountByIDTestSuite) TestMissingAuthorizationHeader() {
	suite.T().Run("missing authorization header returns 401", func(t *testing.T) {
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts/12345678901234", http.MethodGet, nil, nil)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Authorization header is missing")
	})
}

func (suite *GetAccountByIDTestSuite) TestInvalidAuthorizationHeaderFormat() {
	suite.T().Run("invalid authorization header format returns 401", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "InvalidFormat token123",
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts/12345678901234", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Invalid Authorization header format")
	})
}

func (suite *GetAccountByIDTestSuite) TestInvalidToken() {
	suite.T().Run("invalid token returns 401", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer invalid_token_123",
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts/12345678901234", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Invalid or expired token")
	})
}

func (suite *GetAccountByIDTestSuite) TestExpiredToken() {
	suite.T().Run("expired token returns 401", func(t *testing.T) {
		expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjUwNDAwMDAsInVzZXJfaWQiOiJhMWIyYzNkNC1lNWY2LTRhNWItOGM5ZC0wZTFmMmEzYjRjNWQifQ.invalid"

		headers := map[string]string{
			"Authorization": "Bearer " + expiredToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts/12345678901234", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Invalid or expired token")
	})
}

func (suite *GetAccountByIDTestSuite) TestInvalidAccountID() {
	suite.T().Run("invalid account ID format returns 400", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts/invalid_id", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Invalid account ID")
	})
}

func (suite *GetAccountByIDTestSuite) TestAccountNotFound() {
	suite.T().Run("non-existent account ID returns 404", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		// use a non-existent account ID
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts/99999999999999", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Account not found")
	})
}

func (suite *GetAccountByIDTestSuite) TestUserCannotAccessOthersAccount() {
	suite.T().Run("user cannot access another user's account", func(t *testing.T) {
		// user 1 tries to access user 2's account
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		// try to access account belonging to user 2
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts/11111111111111", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusForbidden, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "You do not have permission to access this account")
	})
}

func (suite *GetAccountByIDTestSuite) TestSuccessfulGetAccountByID() {
	suite.T().Run("valid request returns account details", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts/12345678901234", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response types.GetAccountByIDResponse
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		// verify account details
		assert.Equal(t, int64(12345678901234), response.Data.ID)
		assert.Equal(t, userID, response.Data.UserID)
		assert.Equal(t, int64(150000), response.Data.Balance)
		assert.Equal(t, "SAVINGS_ACCOUNT", string(response.Data.Type))
		assert.NotZero(t, response.Data.CreatedAt)
		assert.NotZero(t, response.Data.UpdatedAt)
	})
}

func (suite *GetAccountByIDTestSuite) TestResponseFormat() {
	suite.T().Run("response has correct format", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts/12345678901234", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response map[string]interface{}
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		// check that response has "data" field
		data, exists := response["data"]
		assert.True(t, exists, "response should have 'data' field")
		assert.NotNil(t, data)

		// verify data is an object (not an array)
		dataObject, ok := data.(map[string]interface{})
		assert.True(t, ok, "data should be an object")

		// verify all required fields exist
		requiredFields := []string{"id", "created_at", "updated_at", "user_id", "balance", "type"}
		for _, field := range requiredFields {
			_, exists := dataObject[field]
			assert.True(t, exists, "account should contain field: %s", field)
		}
	})
}

func (suite *GetAccountByIDTestSuite) TestDifferentAccountTypes() {
	suite.T().Run("can fetch CURRENT_ACCOUNT", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		// fetch CURRENT_ACCOUNT
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts/98765432109876", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response types.GetAccountByIDResponse
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, int64(98765432109876), response.Data.ID)
		assert.Equal(t, "CURRENT_ACCOUNT", string(response.Data.Type))
		assert.Equal(t, int64(98000), response.Data.Balance)
	})

	suite.T().Run("can fetch SAVINGS_ACCOUNT", func(t *testing.T) {
		userID := "b2c3d4e5-f6a7-5b6c-9d0e-1f2a3b4c5d6e"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		// fetch SAVINGS_ACCOUNT
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/accounts/11111111111111", http.MethodGet, nil, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response types.GetAccountByIDResponse
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, int64(11111111111111), response.Data.ID)
		assert.Equal(t, "SAVINGS_ACCOUNT", string(response.Data.Type))
		assert.Equal(t, int64(50000), response.Data.Balance)
	})
}
