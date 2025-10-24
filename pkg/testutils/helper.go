package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ErrorResponse struct {
	Error struct {
		StatusCode int `json:"status_code"`
		Details    any `json:"details"`
	} `json:"error"`
}

func DecodeErrorResponse(t *testing.T, w *httptest.ResponseRecorder) ErrorResponse {
	t.Helper()

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	return response
}

func MakeRequest(t *testing.T, app TestApp, endpoint string, httpMethod string, requestPayload any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody *bytes.Buffer
	if requestPayload != nil {
		body, err := json.Marshal(requestPayload)
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
		requestBody = bytes.NewBuffer(body)
	} else {
		requestBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(httpMethod, endpoint, requestBody)
	assert.Equal(t, nil, err)

	// add customheaders
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	return w
}

func AssertFieldError(t *testing.T, resp ErrorResponse, field string, expectedErrMsg string) {
	t.Helper()

	details, ok := resp.Error.Details.(map[string]any)
	if !ok {
		t.Fatalf("expected details to be a map, got %T", resp.Error.Details)
	}

	actualErrMsg, exists := details[field]
	assert.Equal(t, true, exists)
	assert.Equal(t, expectedErrMsg, actualErrMsg)
}
