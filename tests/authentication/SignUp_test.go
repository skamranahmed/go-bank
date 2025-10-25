package authentication

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	accountModel "github.com/skamranahmed/go-bank/internal/account/model"
	"github.com/skamranahmed/go-bank/internal/authentication/dto"
	userModel "github.com/skamranahmed/go-bank/internal/user/model"

	userTasks "github.com/skamranahmed/go-bank/internal/user/tasks"
	"github.com/skamranahmed/go-bank/mock"
	tasksHelper "github.com/skamranahmed/go-bank/pkg/tasks"
	"github.com/skamranahmed/go-bank/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// Define your test suite struct
type SignUpTestSuite struct {
	suite.Suite
	app testutils.TestApp
}

// This function actually runs the test suite
func TestSignUpTestSuite(t *testing.T) {
	suite.Run(t, new(SignUpTestSuite))
}

// SetupSuite runs once before all tests
func (suite *SignUpTestSuite) SetupSuite() {
	// setup app
	suite.app = testutils.NewTestApp(suite.T().Context(), nil, postgresTestContainer, redisTestContainer)

	// load fixtures
	fixtures, err := testfixtures.New(
		testfixtures.Database(suite.app.Db.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("./fixtures/SignUp_test"),
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
func (suite *SignUpTestSuite) TearDownSuite() {
	suite.app.TeardownFunc()
}

func (suite *SignUpTestSuite) TestValidationErrors() {
	type scenario struct {
		name               string
		payload            dto.SignUpRequest
		field              string
		errMessage         string
		expectedStatusCode int
	}

	tests := []scenario{
		{
			name: "missing email",
			payload: dto.SignUpRequest{
				Data: dto.SignUpData{
					Username: "username",
					Password: "password",
				},
			},
			field:              "email",
			errMessage:         "email is a required field",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "empty email",
			payload: dto.SignUpRequest{
				Data: dto.SignUpData{
					Email:    "",
					Username: "username",
					Password: "password",
				},
			},
			field:              "email",
			errMessage:         "email is a required field",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid email",
			payload: dto.SignUpRequest{
				Data: dto.SignUpData{
					Email:    "not_an_email",
					Username: "username",
					Password: "password",
				},
			},
			field:              "email",
			errMessage:         "not_an_email is not a valid email",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "missing username",
			payload: dto.SignUpRequest{
				Data: dto.SignUpData{
					Email:    "test_user_1@example.com",
					Password: "password",
				},
			},
			field:              "username",
			errMessage:         "username is a required field",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "empty username",
			payload: dto.SignUpRequest{
				Data: dto.SignUpData{
					Email:    "test_user_1@example.com",
					Username: "",
					Password: "password",
				},
			},
			field:              "username",
			errMessage:         "username is a required field",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "short username (less than 8 characters)",
			payload: dto.SignUpRequest{
				Data: dto.SignUpData{
					Email:    "test_user_1@example.com",
					Username: "user",
					Password: "password",
				},
			},
			field:              "username",
			errMessage:         "username must be at least 8 characters",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "missing password",
			payload: dto.SignUpRequest{
				Data: dto.SignUpData{
					Email:    "test_user_1@example.com",
					Username: "username",
				},
			},
			field:              "password",
			errMessage:         "password is a required field",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "empty password",
			payload: dto.SignUpRequest{
				Data: dto.SignUpData{
					Email:    "test_user_1@example.com",
					Username: "username",
					Password: "",
				},
			},
			field:              "password",
			errMessage:         "password is a required field",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "short password (less than 8 characters)",
			payload: dto.SignUpRequest{
				Data: dto.SignUpData{
					Email:    "test_user_1@example.com",
					Username: "username",
					Password: "pass",
				},
			},
			field:              "password",
			errMessage:         "password must be at least 8 characters",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/sign-up", http.MethodPost, tc.payload, nil)
			assert.Equal(t, tc.expectedStatusCode, responseRecorder.Code)

			response := testutils.DecodeErrorResponse(t, responseRecorder)
			testutils.AssertFieldError(t, response, tc.field, tc.errMessage)
		})
	}

}

func (suite *SignUpTestSuite) TestConflictErrors() {
	type scenario struct {
		name               string
		payload            dto.SignUpRequest
		field              string
		errMessage         string
		expectedStatusCode int
	}

	tests := []scenario{
		{
			name: "duplicate email",
			payload: dto.SignUpRequest{
				Data: dto.SignUpData{
					Email:    "kamran@example.com",
					Username: "username",
					Password: "password",
				},
			},
			field:              "message",
			errMessage:         "This username or email is already in use. Please choose another.",
			expectedStatusCode: http.StatusConflict,
		},
		{
			name: "duplicate username",
			payload: dto.SignUpRequest{
				Data: dto.SignUpData{
					Email:    "test_user_1@example.com",
					Username: "kamran_ahmed",
					Password: "password",
				},
			},
			field:              "message",
			errMessage:         "This username or email is already in use. Please choose another.",
			expectedStatusCode: http.StatusConflict,
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			responseRecorder := testutils.MakeRequest(t, suite.app, "/v1/sign-up", http.MethodPost, tc.payload, nil)
			assert.Equal(t, tc.expectedStatusCode, responseRecorder.Code)

			response := testutils.DecodeErrorResponse(t, responseRecorder)
			testutils.AssertFieldError(t, response, tc.field, tc.errMessage)
		})
	}
}

func (suite *SignUpTestSuite) TestSuccessfulSignUp() {
	type SuccessResponse struct {
		AccessToken string `json:"access_token"`
	}

	suite.T().Run("successful signup creates user record, account record, access token, and enqueues task", func(t *testing.T) {
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		// setup expectations for task enqueuing
		var enqueuedTask tasksHelper.Task
		mockTaskEnqueuer := mock.NewMockTaskEnqueuer(mockController)
		mockTaskEnqueuer.EXPECT().
			Enqueue(gomock.Any(), gomock.Any(), nil, nil).
			Do(func(ctx context.Context, task tasksHelper.Task, maxRetryCount *int, queueName *string) {
				enqueuedTask = task
			}).
			Return(nil).
			Times(1)

		appWithMock := testutils.NewTestApp(
			suite.T().Context(),
			&testutils.TestAppDeps{
				Db:           suite.app.Db,     // reuse the db from the app
				Cache:        suite.app.Cache,  // reuse the cache from the app
				TaskEnqueuer: mockTaskEnqueuer, // inject mock TaskEnqueuer to verify task enqueuing
			},
			nil,
			nil,
		)

		// prepare request payload
		payload := dto.SignUpRequest{
			Data: dto.SignUpData{
				Email:    "test_user_1@example.com",
				Username: "username",
				Password: "password",
			},
		}

		responseRecorder := testutils.MakeRequest(t, appWithMock, "/v1/sign-up", http.MethodPost, payload, nil)
		assert.Equal(t, http.StatusCreated, responseRecorder.Code)

		var response SuccessResponse
		err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.NotEmpty(t, response.AccessToken)

		// check user record
		var user userModel.User
		err = appWithMock.Db.NewSelect().
			Model(&user).
			Where("email = ?", "test_user_1@example.com").
			Limit(1).
			Scan(t.Context())

		// no error
		assert.Equal(t, nil, err)

		// assert user record data
		assert.NotZero(t, user.ID)
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)
		assert.NotZero(t, user.Password)
		assert.Equal(t, "test_user_1@example.com", user.Email)
		assert.Equal(t, "username", user.Username)
		assert.NotEqual(t, "password", user.Password) // must not match the plain text password provided by the user

		// check account record
		var account accountModel.Account
		err = appWithMock.Db.NewSelect().
			Model(&account).
			Where("user_id = ?", user.ID).
			Limit(1).
			Scan(t.Context())

		// no error
		assert.Equal(t, nil, err)

		// assert account record data
		assert.NotZero(t, account.ID)
		assert.NotZero(t, account.CreatedAt)
		assert.NotZero(t, account.UpdatedAt)
		assert.Equal(t, int64(0), account.Balance)
		assert.Equal(t, accountModel.SavingsAccount, account.Type)

		// verify the access token that is returned in response and check its existence in cache
		tokenData, err := appWithMock.Services.AuthenticationService.VerifyAccessToken(t.Context(), response.AccessToken)
		assert.Equal(t, nil, err)
		assert.NotZero(t, tokenData.UserID)
		assert.NotZero(t, tokenData.TokenID)

		assert.Equal(t, user.ID.String(), tokenData.UserID)

		accessTokenCacheKey := fmt.Sprintf("auth:access_token_id:%s:user_id:%s", tokenData.TokenID, user.ID)
		tokenInCache, err := appWithMock.Cache.Get(t.Context(), accessTokenCacheKey)
		assert.Equal(t, nil, err)
		assert.Equal(t, "", tokenInCache)

		// assert enqueued task details
		assert.Equal(t, userTasks.SendWelcomeEmailTaskName, enqueuedTask.Name())
		enqueuedTaskPayload, ok := enqueuedTask.Payload().(userTasks.SendWelcomeEmailTaskPayload)
		assert.Equal(t, true, ok)
		assert.Equal(t, user.ID.String(), enqueuedTaskPayload.UserID)
	})
}
