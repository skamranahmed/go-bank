package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	accountModel "github.com/skamranahmed/go-bank/internal/account/model"
	"github.com/skamranahmed/go-bank/internal/transfer/types"
	"github.com/skamranahmed/go-bank/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Helper function to create pointer to int64
func int64Ptr(i int64) *int64 {
	return &i
}

type PerformInternalTransferTestSuite struct {
	suite.Suite
	app testutils.TestApp
}

func TestPerformInternalTransferTestSuite(t *testing.T) {
	suite.Run(t, new(PerformInternalTransferTestSuite))
}

// SetupSuite runs once before all tests
func (suite *PerformInternalTransferTestSuite) SetupSuite() {
	suite.app = testutils.NewTestApp(suite.T().Context(), nil, postgresTestContainer, redisTestContainer)

	fixtures, err := testfixtures.New(
		testfixtures.Database(suite.app.Db.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("./fixtures/PerformInternalTransfer_test"),
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
func (suite *PerformInternalTransferTestSuite) TearDownSuite() {
	suite.app.TeardownFunc()
}

func (suite *PerformInternalTransferTestSuite) TestMissingAuthorizationHeader() {
	suite.T().Run("missing authorization header returns 401", func(t *testing.T) {
		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 12345678901234,
				ToAccountID:   11111111111111,
				Amount:        int64Ptr(10000),
			},
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, nil)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Authorization header is missing")
	})
}

func (suite *PerformInternalTransferTestSuite) TestInvalidAuthorizationHeaderFormat() {
	suite.T().Run("invalid authorization header format returns 401", func(t *testing.T) {
		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 12345678901234,
				ToAccountID:   11111111111111,
				Amount:        int64Ptr(10000),
			},
		}

		headers := map[string]string{
			"Authorization": "InvalidFormat token123",
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Invalid Authorization header format")
	})
}

func (suite *PerformInternalTransferTestSuite) TestInvalidToken() {
	suite.T().Run("invalid token returns 401", func(t *testing.T) {
		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 12345678901234,
				ToAccountID:   11111111111111,
				Amount:        int64Ptr(10000),
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer invalid_token_123",
		}
		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
		assert.Equal(t, http.StatusUnauthorized, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Invalid or expired token")
	})
}

func (suite *PerformInternalTransferTestSuite) TestValidationErrors() {
	type scenario struct {
		name               string
		payload            types.InternalTransferRequest
		field              string
		errMessage         string
		expectedStatusCode int
	}

	tests := []scenario{
		{
			name: "missing from_account_id",
			payload: types.InternalTransferRequest{
				Data: types.InternalTransferRequestData{
					ToAccountID: 11111111111111,
					Amount:      int64Ptr(10000),
				},
			},
			field:              "from_account_id",
			errMessage:         "from_account_id is a required field",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "missing to_account_id",
			payload: types.InternalTransferRequest{
				Data: types.InternalTransferRequestData{
					FromAccountID: 12345678901234,
					Amount:        int64Ptr(10000),
				},
			},
			field:              "to_account_id",
			errMessage:         "to_account_id is a required field",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "missing amount",
			payload: types.InternalTransferRequest{
				Data: types.InternalTransferRequestData{
					FromAccountID: 12345678901234,
					ToAccountID:   11111111111111,
				},
			},
			field:              "amount",
			errMessage:         "amount is a required field",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "zero amount",
			payload: types.InternalTransferRequest{
				Data: types.InternalTransferRequestData{
					FromAccountID: 12345678901234,
					ToAccountID:   11111111111111,
					Amount:        int64Ptr(0),
				},
			},
			field:              "amount",
			errMessage:         "amount must be greater than 0",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "negative amount",
			payload: types.InternalTransferRequest{
				Data: types.InternalTransferRequestData{
					FromAccountID: 12345678901234,
					ToAccountID:   11111111111111,
					Amount:        int64Ptr(-5000),
				},
			},
			field:              "amount",
			errMessage:         "amount must be greater than 0",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

			accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
			assert.NoError(t, err)

			headers := map[string]string{
				"Authorization": "Bearer " + accessToken,
			}

			responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, tc.payload, headers)
			assert.Equal(t, tc.expectedStatusCode, responseRecorder.Code)

			response := testutils.DecodeErrorResponse(t, responseRecorder)
			testutils.AssertFieldError(t, response, tc.field, tc.errMessage)
		})
	}
}

func (suite *PerformInternalTransferTestSuite) TestSameAccountTransfer() {
	suite.T().Run("transfer to same account returns 400", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 12345678901234,
				ToAccountID:   12345678901234, // same as from_account_id
				Amount:        int64Ptr(10000),
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
		assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Sender and recipient account ids must be different")
	})
}

func (suite *PerformInternalTransferTestSuite) TestFromAccountNotFound() {
	suite.T().Run("non-existent from_account_id returns 404", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 99999999999999, // non-existent account
				ToAccountID:   11111111111111,
				Amount:        int64Ptr(10000),
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
		assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Account not found")
	})
}

func (suite *PerformInternalTransferTestSuite) TestUnauthorizedFromAccount() {
	suite.T().Run("user cannot transfer from another user's account", func(t *testing.T) {
		// user 1 tries to transfer from user 2's account
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 11111111111111, // belongs to user 2
				ToAccountID:   12345678901234, // belongs to user 1
				Amount:        int64Ptr(10000),
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
		assert.Equal(t, http.StatusForbidden, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "You are not authorized to perform transfer from this account")
	})
}

func (suite *PerformInternalTransferTestSuite) TestToAccountNotFound() {
	suite.T().Run("non-existent to_account_id returns 404", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 12345678901234,
				ToAccountID:   99999999999999, // non-existent account
				Amount:        int64Ptr(10000),
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
		assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "Account not found")
	})
}

func (suite *PerformInternalTransferTestSuite) TestInsufficientBalance() {
	suite.T().Run("insufficient balance returns 400", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 12345678901234, // has balance of 150000
				ToAccountID:   11111111111111,
				Amount:        int64Ptr(200000), // more than available balance
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
		assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

		response := testutils.DecodeErrorResponse(t, responseRecorder)
		testutils.AssertFieldError(t, response, "message", "You do not have sufficient balance in your account to perform the transfer")
	})
}

func (suite *PerformInternalTransferTestSuite) TestSuccessfulTransfer() {
	suite.T().Run("successful transfer creates transaction records and updates balances", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		// get initial balances
		var senderAccountBefore accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&senderAccountBefore).
			Where("id = ?", 12345678901234).
			Scan(t.Context())
		assert.NoError(t, err)

		var receiverAccountBefore accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&receiverAccountBefore).
			Where("id = ?", 11111111111111).
			Scan(t.Context())
		assert.NoError(t, err)

		transferAmount := int64(25000)
		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 12345678901234,
				ToAccountID:   11111111111111,
				Amount:        &transferAmount,
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response types.InternalTransferResponse
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		// verify transaction in response
		assert.NotEmpty(t, response.Data.Transaction.ID)
		assert.Equal(t, int64(12345678901234), response.Data.Transaction.AccountID)
		assert.Equal(t, transferAmount, response.Data.Transaction.Amount)
		assert.Equal(t, string(accountModel.Debit), response.Data.Transaction.Type)
		assert.Equal(t, senderAccountBefore.Balance-transferAmount, response.Data.Transaction.BalanceAfter)
		assert.NotZero(t, response.Data.Transaction.CreatedAt)

		// verify sender account balance updated
		var senderAccountAfter accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&senderAccountAfter).
			Where("id = ?", 12345678901234).
			Scan(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, senderAccountBefore.Balance-transferAmount, senderAccountAfter.Balance)

		// verify receiver account balance updated
		var receiverAccountAfter accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&receiverAccountAfter).
			Where("id = ?", 11111111111111).
			Scan(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, receiverAccountBefore.Balance+transferAmount, receiverAccountAfter.Balance)

		// verify transaction record for sender account exists
		var senderTransaction accountModel.Transaction
		err = suite.app.Db.NewSelect().
			Model(&senderTransaction).
			Where("account_id = ? AND type = ?", 12345678901234, accountModel.Debit).
			Order("created_at DESC").
			Limit(1).
			Scan(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, transferAmount, senderTransaction.Amount)
		assert.Equal(t, accountModel.Debit, senderTransaction.Type)
		assert.Equal(t, senderAccountAfter.Balance, senderTransaction.BalanceAfter)

		// verify transaction record for receiver account exists
		var receiverTransaction accountModel.Transaction
		err = suite.app.Db.NewSelect().
			Model(&receiverTransaction).
			Where("account_id = ? AND type = ?", 11111111111111, accountModel.Credit).
			Order("created_at DESC").
			Limit(1).
			Scan(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, transferAmount, receiverTransaction.Amount)
		assert.Equal(t, accountModel.Credit, receiverTransaction.Type)
		assert.Equal(t, receiverAccountAfter.Balance, receiverTransaction.BalanceAfter)
	})
}

func (suite *PerformInternalTransferTestSuite) TestTransferBetweenDifferentAccountTypes() {
	suite.T().Run("transfer from SAVINGS_ACCOUNT to CURRENT_ACCOUNT", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		transferAmount := int64(10000)
		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 12345678901234, // SAVINGS_ACCOUNT
				ToAccountID:   98765432109876, // CURRENT_ACCOUNT
				Amount:        &transferAmount,
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response types.InternalTransferResponse
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Data.Transaction.ID)
	})

	suite.T().Run("transfer from CURRENT_ACCOUNT to SAVINGS_ACCOUNT", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		transferAmount := int64(5000)
		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 98765432109876, // CURRENT_ACCOUNT
				ToAccountID:   12345678901234, // SAVINGS_ACCOUNT
				Amount:        &transferAmount,
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response types.InternalTransferResponse
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Data.Transaction.ID)
	})
}

func (suite *PerformInternalTransferTestSuite) TestTransferExactBalance() {
	suite.T().Run("transfer exact balance empties sender account", func(t *testing.T) {
		userID := "c3d4e5f6-a7b8-6c7d-0e1f-2a3b4c5d6e7f"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		// get sender account with low balance
		var senderAccountBefore accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&senderAccountBefore).
			Where("id = ?", 22222222222222).
			Scan(t.Context())
		assert.NoError(t, err)

		transferAmount := senderAccountBefore.Balance // transfer exact balance
		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 22222222222222,
				ToAccountID:   11111111111111,
				Amount:        &transferAmount,
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
		assert.Equal(t, http.StatusOK, responseRecorder.Code)

		var response types.InternalTransferResponse
		err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		// verify sender account has zero balance
		var senderAccountAfter accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&senderAccountAfter).
			Where("id = ?", 22222222222222).
			Scan(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, int64(0), senderAccountAfter.Balance)
		assert.Equal(t, int64(0), response.Data.Transaction.BalanceAfter)
	})
}

func (suite *PerformInternalTransferTestSuite) TestResponseFormat() {
	suite.T().Run("response has correct format", func(t *testing.T) {
		userID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		transferAmount := int64(1000)
		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 12345678901234,
				ToAccountID:   11111111111111,
				Amount:        &transferAmount,
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
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

		// verify transaction field exists
		transaction, exists := dataObject["transaction"]
		assert.True(t, exists, "data should contain 'transaction' field")
		assert.NotNil(t, transaction)

		// verify transaction fields
		transactionObject, ok := transaction.(map[string]interface{})
		assert.True(t, ok, "transaction should be an object")

		requiredFields := []string{"id", "created_at", "account_id", "amount", "type", "balance_after"}
		for _, field := range requiredFields {
			_, exists := transactionObject[field]
			assert.True(t, exists, "transaction should contain field: %s", field)
		}
	})
}

func (suite *PerformInternalTransferTestSuite) TestConcurrentTransfersFromSameSenderAccount() {
	suite.T().Run("should correctly update the balance of sender's and receiver's account after all tranfers have completed", func(t *testing.T) {
		userID := "d3d4e5f6-a7b8-6c7d-0e1f-2a3b4c5d6e7f"

		accessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), userID)
		assert.NoError(t, err)

		transferAmount := int64(10000)
		payload := types.InternalTransferRequest{
			Data: types.InternalTransferRequestData{
				FromAccountID: 33333333333333,
				ToAccountID:   44444444444444,
				Amount:        &transferAmount,
			},
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		// get sender's account balance before
		var senderAccountBefore accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&senderAccountBefore).
			Where("id = ?", 33333333333333).
			Scan(t.Context())
		assert.NoError(t, err)
		senderAccountBalanceBefore := senderAccountBefore.Balance

		// get receiver's account balance before
		var receiverAccountBefore accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&receiverAccountBefore).
			Where("id = ?", 44444444444444).
			Scan(t.Context())
		assert.NoError(t, err)
		receiverAccountBalanceBefore := receiverAccountBefore.Balance

		// 50 concurrent transfers of 10000 each, total debit amount = 50 * 10000 = 500000
		numOfConcurrentTransfers := 50
		errChan := make(chan error, numOfConcurrentTransfers)

		var wg sync.WaitGroup

		for i := 0; i < numOfConcurrentTransfers; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
				var response map[string]interface{}
				err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				if responseRecorder.Code != http.StatusOK {
					errMsg := fmt.Sprintf("expected status code %d, got %d, response %+v", http.StatusOK, responseRecorder.Code, response)
					errChan <- errors.New(errMsg)
				}

				if err != nil {
					errChan <- err
				}
			}()
		}

		// wait for all goroutines to finish
		wg.Wait()

		// close the error channel and collect all errors
		close(errChan)
		for err := range errChan {
			t.Error(err)
		}

		// verify sender's account balance after
		senderAccountBalanceAfter := senderAccountBalanceBefore - int64(numOfConcurrentTransfers)*transferAmount
		var senderAccountAfter accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&senderAccountAfter).
			Where("id = ?", 33333333333333).
			Scan(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, senderAccountBalanceAfter, senderAccountAfter.Balance)

		// verify receiver's account balance after
		receiverAccountBalanceAfter := receiverAccountBalanceBefore + int64(numOfConcurrentTransfers)*transferAmount
		var receiverAccountAfter accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&receiverAccountAfter).
			Where("id = ?", 44444444444444).
			Scan(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, receiverAccountBalanceAfter, receiverAccountAfter.Balance)
	})
}

func (suite *PerformInternalTransferTestSuite) TestConcurrentTransfersInvolvingSameAccountsWithDifferentRoles() {
	suite.T().Run("should correctly update the balance of sender's and receiver's account after all tranfers have completed", func(t *testing.T) {
		firstUserID := "f3d4e5f6-a7b8-6c7d-0e1f-2a3b4c5d6e7f"
		firstUserAccountID := 55555555555555
		transferAmountFromFirstAccountInEachTransfer := 10000
		firstUserAccessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), firstUserID)
		assert.NoError(t, err)

		// get first user's account balance before
		var firstUserAccountBefore accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&firstUserAccountBefore).
			Where("id = ?", firstUserAccountID).
			Scan(t.Context())
		assert.NoError(t, err)
		firstUserAccountBalanceBefore := firstUserAccountBefore.Balance

		secondUserID := "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5e"
		secondUserAccountID := 66666666666666
		transferAmountFromSecondAccountInEachTransfer := 20000
		secondUserAccessToken, err := suite.app.Services.AuthenticationService.CreateAccessToken(t.Context(), secondUserID)
		assert.NoError(t, err)

		// get second user's account balance before
		var secondUserAccountBefore accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&secondUserAccountBefore).
			Where("id = ?", secondUserAccountID).
			Scan(t.Context())
		assert.NoError(t, err)
		secondUserAccountBalanceBefore := secondUserAccountBefore.Balance

		numOfConcurrentTransfers := 50 // 25 transfers from first account and 25 transfers from second account
		errChan := make(chan error, numOfConcurrentTransfers)

		var wg sync.WaitGroup

		for i := 0; i < numOfConcurrentTransfers; i++ {
			wg.Add(1)

			go func(i int) {
				defer wg.Done()

				var accessToken string
				var senderAccountID, receiverAccountID int64
				var transferAmount int64

				if i%2 == 0 {
					// first user id is the sender and second user id is the receiver
					accessToken = firstUserAccessToken
					senderAccountID = int64(firstUserAccountID)
					receiverAccountID = int64(secondUserAccountID)

					// transfer 10000 in each transfer from first account, i.e total debit = 25 * 10000 = 250000
					transferAmount = int64(transferAmountFromFirstAccountInEachTransfer)
				} else {
					// second user id is the sender and first user id is the receiver
					accessToken = secondUserAccessToken
					senderAccountID = int64(secondUserAccountID)
					receiverAccountID = int64(firstUserAccountID)

					// transfer 20000 in each transfer from second account, i.e total debit = 25 * 10000 = 500000
					transferAmount = int64(transferAmountFromSecondAccountInEachTransfer)
				}

				payload := types.InternalTransferRequest{
					Data: types.InternalTransferRequestData{
						FromAccountID: senderAccountID,
						ToAccountID:   receiverAccountID,
						Amount:        &transferAmount,
					},
				}

				headers := map[string]string{
					"Authorization": "Bearer " + accessToken,
				}

				responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/transfers/internal", http.MethodPost, payload, headers)
				var response map[string]interface{}
				err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				if responseRecorder.Code != http.StatusOK {
					errMsg := fmt.Sprintf("expected status code %d, got %d, response %+v", http.StatusOK, responseRecorder.Code, response)
					errChan <- errors.New(errMsg)
				}

				if err != nil {
					errChan <- err
				}
			}(i)
		}

		// wait for all goroutines to finish
		wg.Wait()

		// close the error channel and collect all errors
		close(errChan)
		for err := range errChan {
			t.Error(err)
		}

		// verify first user's account balance after
		amountDebitedFromFirstUserAccount := int64(25 * transferAmountFromFirstAccountInEachTransfer) // 25 transfers of 10000 each
		amountCreditedToFirstUserAccount := int64(25 * transferAmountFromSecondAccountInEachTransfer) // 25 transfers of 20000 each
		firstUserAccountBalanceAfter := firstUserAccountBalanceBefore - amountDebitedFromFirstUserAccount + amountCreditedToFirstUserAccount
		var firstUserAccountAfter accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&firstUserAccountAfter).
			Where("id = ?", firstUserAccountID).
			Scan(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, firstUserAccountBalanceAfter, firstUserAccountAfter.Balance)

		// verify receiver's account balance after
		amountDebitedFromSecondUserAccount := int64(25 * transferAmountFromSecondAccountInEachTransfer) // 25 transfers of 20000 each
		amountCreditedToSecondUserAccount := int64(25 * transferAmountFromFirstAccountInEachTransfer)   // 25 transfers of 10000 each
		secondUserAccountBalanceAfter := secondUserAccountBalanceBefore - amountDebitedFromSecondUserAccount + amountCreditedToSecondUserAccount
		var receiverAccountAfter accountModel.Account
		err = suite.app.Db.NewSelect().
			Model(&receiverAccountAfter).
			Where("id = ?", secondUserAccountID).
			Scan(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, secondUserAccountBalanceAfter, receiverAccountAfter.Balance)
	})
}
