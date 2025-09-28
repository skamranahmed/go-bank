package healthz

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skamranahmed/go-bank/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func Test_CheckHealth(t *testing.T) {
	ctx := context.TODO()

	t.Run("when database is down: it should return 500 status code and a response body with status: DB_NOT_OK", func(t *testing.T) {
		app := testutils.NewTestApp(ctx, nil, postgresTestContainer, redisTestContainer)
		defer app.TeardownFunc()

		// close the db connection to simulate db down behaviour
		app.Db.Close()

		req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		type Response struct {
			Status string `json:"status"`
		}

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Equal(t, "DB_NOT_OK", response.Status)
	})

	t.Run("when cache is down: it should return 500 status code and a response body with status: CACHE_NOT_OK", func(t *testing.T) {
		app := testutils.NewTestApp(ctx, nil, postgresTestContainer, redisTestContainer)
		defer app.TeardownFunc()

		// close the cache connection to simulate cache down behaviour
		app.Cache.Close()

		req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		type Response struct {
			Status string `json:"status"`
		}

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Equal(t, "CACHE_NOT_OK", response.Status)
	})

	t.Run("when everything is up and running: it should return 200 status code and a response body with status: ALL_OK", func(t *testing.T) {
		app := testutils.NewTestApp(ctx, nil, postgresTestContainer, redisTestContainer)
		defer app.TeardownFunc()

		req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		type Response struct {
			Status string `json:"status"`
		}

		var response Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		assert.Equal(t, "ALL_OK", response.Status)
	})
}
