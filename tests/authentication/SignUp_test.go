package authentication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	accountModel "github.com/skamranahmed/go-bank/internal/account/model"
	"github.com/skamranahmed/go-bank/internal/authentication/dto"
	userModel "github.com/skamranahmed/go-bank/internal/user/model"
	"github.com/skamranahmed/go-bank/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

type ErrorResponse struct {
	Error struct {
		StatusCode int `json:"status_code"`
		Details    any `json:"details"`
	} `json:"error"`
}

type SuccessResponse struct {
	AccessToken string `json:"access_token"`
}

func makeRequest(t *testing.T, app testutils.TestApp, requestPayload any) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(requestPayload)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/v1/sign-up", bytes.NewBuffer(body))
	assert.Equal(t, nil, err)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	return w
}

func decodeErrorResponse(t *testing.T, w *httptest.ResponseRecorder) ErrorResponse {
	t.Helper()

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	return response
}

func assertFieldError(t *testing.T, resp ErrorResponse, field string, expectedErrMsg string) {
	t.Helper()

	details, ok := resp.Error.Details.(map[string]any)
	if !ok {
		t.Fatalf("expected details to be a map, got %T", resp.Error.Details)
	}

	actualErrMsg, exists := details[field]
	assert.Equal(t, true, exists)
	assert.Equal(t, expectedErrMsg, actualErrMsg)
}

func Test_SignUp_Route(t *testing.T) {
	ctx := context.TODO()

	app := testutils.NewApp(ctx, postgresTestContainer, redisTestContainer)
	defer app.TeardownFunc()

	fixtures, err := testfixtures.New(
		testfixtures.Database(app.Db.DB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("./fixtures/SignUp_test"),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = fixtures.Load()
	if err != nil {
		t.Fatal(err)
	}

	// unhappy paths
	t.Run("when email is NOT provided in the request payload: it should return 400 status code", func(t *testing.T) {
		// prepare payload
		reqBody := dto.SignUpRequest{}
		reqBody.Data.Username = "username" // 8 chars
		reqBody.Data.Password = "password" // 8 chars

		w := makeRequest(t, app, reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		response := decodeErrorResponse(t, w)
		assert.Equal(t, http.StatusBadRequest, response.Error.StatusCode)
		assertFieldError(t, response, "email", "email is a required field")
	})

	t.Run("when empty email string is provided in the request payload: it should return 400 status code", func(t *testing.T) {
		// prepare payload
		reqBody := dto.SignUpRequest{}
		reqBody.Data.Email = ""
		reqBody.Data.Username = "username" // 8 chars
		reqBody.Data.Password = "password" // 8 chars

		w := makeRequest(t, app, reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		response := decodeErrorResponse(t, w)
		assert.Equal(t, http.StatusBadRequest, response.Error.StatusCode)
		assertFieldError(t, response, "email", "email is a required field")
	})

	t.Run("when an invalid email is provided in the request payload: it should return 400 status code", func(t *testing.T) {
		// prepare payload
		reqBody := dto.SignUpRequest{}
		reqBody.Data.Email = "not_an_email" // invalid email
		reqBody.Data.Username = "username"  // 8 chars
		reqBody.Data.Password = "password"  // 8 chars

		w := makeRequest(t, app, reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		response := decodeErrorResponse(t, w)
		assert.Equal(t, http.StatusBadRequest, response.Error.StatusCode)
		assertFieldError(t, response, "email", "not_an_email is not a valid email")
	})

	t.Run("when username is NOT provided in the request payload: it should return 400 status code", func(t *testing.T) {
		// prepare payload
		reqBody := dto.SignUpRequest{}
		reqBody.Data.Email = "test_user_1@example.com" // valid email
		reqBody.Data.Password = "password"             // 8 chars

		w := makeRequest(t, app, reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		response := decodeErrorResponse(t, w)
		assert.Equal(t, http.StatusBadRequest, response.Error.StatusCode)
		assertFieldError(t, response, "username", "username is a required field")
	})

	t.Run("when empty username string is provided in the request payload: it should return 400 status code", func(t *testing.T) {
		// prepare payload
		reqBody := dto.SignUpRequest{}
		reqBody.Data.Email = "test_user_1@example.com" // valid email
		reqBody.Data.Username = ""
		reqBody.Data.Password = "password" // 8 chars

		w := makeRequest(t, app, reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		response := decodeErrorResponse(t, w)
		assert.Equal(t, http.StatusBadRequest, response.Error.StatusCode)
		assertFieldError(t, response, "username", "username is a required field")
	})

	t.Run("when a username with less than 8 characters is provided in the request payload: it should return 400 status code", func(t *testing.T) {
		// prepare payload
		reqBody := dto.SignUpRequest{}
		reqBody.Data.Email = "test_user_1@example.com"
		reqBody.Data.Username = "user"     // 4 chars
		reqBody.Data.Password = "password" // 8 chars

		w := makeRequest(t, app, reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		response := decodeErrorResponse(t, w)
		assert.Equal(t, http.StatusBadRequest, response.Error.StatusCode)
		assertFieldError(t, response, "username", "username must be at least 8 characters")
	})

	t.Run("when password is NOT provided in the request payload: it should return 400 status code", func(t *testing.T) {
		// prepare payload
		reqBody := dto.SignUpRequest{}
		reqBody.Data.Email = "test_user_1@example.com" // valid email
		reqBody.Data.Username = "username"             // 8 chars

		w := makeRequest(t, app, reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		response := decodeErrorResponse(t, w)
		assert.Equal(t, http.StatusBadRequest, response.Error.StatusCode)
		assertFieldError(t, response, "password", "password is a required field")
	})

	t.Run("when empty password string is provided in the request payload: it should return 400 status code", func(t *testing.T) {
		// prepare payload
		reqBody := dto.SignUpRequest{}
		reqBody.Data.Email = "test_user_1@example.com" // valid email
		reqBody.Data.Username = "username"             // 8 chars
		reqBody.Data.Password = ""

		w := makeRequest(t, app, reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		response := decodeErrorResponse(t, w)
		assert.Equal(t, http.StatusBadRequest, response.Error.StatusCode)
		assertFieldError(t, response, "password", "password is a required field")
	})

	t.Run("when a password with less than 8 characters is provided in the request payload: it should return 400 status code", func(t *testing.T) {
		// prepare payload
		reqBody := dto.SignUpRequest{}
		reqBody.Data.Email = "test_user_1@example.com"
		reqBody.Data.Username = "username" // 8 chars
		reqBody.Data.Password = "pass"     // 4 chars

		w := makeRequest(t, app, reqBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		response := decodeErrorResponse(t, w)
		assert.Equal(t, http.StatusBadRequest, response.Error.StatusCode)
		assertFieldError(t, response, "password", "password must be at least 8 characters")
	})

	t.Run("when the request payload is correct but a user with the provided email already exists: it should return 409 status code", func(t *testing.T) {
		// prepare payload
		reqBody := dto.SignUpRequest{}
		reqBody.Data.Email = "kamran@example.com"
		reqBody.Data.Username = "username" // 8 chars
		reqBody.Data.Password = "password" // 8 chars

		w := makeRequest(t, app, reqBody)
		assert.Equal(t, http.StatusConflict, w.Code)

		response := decodeErrorResponse(t, w)
		assert.Equal(t, http.StatusConflict, response.Error.StatusCode)
		assertFieldError(t, response, "message", "This username or email is already in use. Please choose another.")
	})

	t.Run("when the request payload is correct but a user with the provided username already exists: it should return 409 status code", func(t *testing.T) {
		// prepare payload
		reqBody := dto.SignUpRequest{}
		reqBody.Data.Email = "test_user_1@example.com"
		reqBody.Data.Username = "kamran_ahmed" // > 8 chars
		reqBody.Data.Password = "password"     // 8 chars

		w := makeRequest(t, app, reqBody)
		assert.Equal(t, http.StatusConflict, w.Code)

		response := decodeErrorResponse(t, w)
		assert.Equal(t, http.StatusConflict, response.Error.StatusCode)
		assertFieldError(t, response, "message", "This username or email is already in use. Please choose another.")
	})

	// happy path
	t.Run("when the request payload is correct: it should return 200 status code and the user record, account record and access token must be created", func(t *testing.T) {
		// prepare payload
		reqBody := dto.SignUpRequest{}
		reqBody.Data.Email = "test_user_1@example.com"
		reqBody.Data.Username = "username" // 8 chars
		reqBody.Data.Password = "password" // 8 chars

		w := makeRequest(t, app, reqBody)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response SuccessResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.NotEmpty(t, response.AccessToken)

		// check user record
		var user userModel.User
		err = app.Db.NewSelect().
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
		err = app.Db.NewSelect().
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

		// verify the access token that is returned in response and check its existence in cache
		tokenData, err := app.Services.AuthenticationService.VerifyAccessToken(t.Context(), response.AccessToken)
		assert.Equal(t, nil, err)
		assert.NotZero(t, tokenData.UserID)
		assert.NotZero(t, tokenData.TokenID)

		assert.Equal(t, user.ID.String(), tokenData.UserID)

		accessTokenCacheKey := fmt.Sprintf("auth:access_token_id:%s:user_id:%s", tokenData.TokenID, user.ID)
		tokenInCache, err := app.Cache.Get(t.Context(), accessTokenCacheKey)
		assert.Equal(t, nil, err)
		assert.Equal(t, "", tokenInCache)
	})
}
