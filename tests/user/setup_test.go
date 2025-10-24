package user

import (
	"context"
	"os"
	"testing"

	"github.com/skamranahmed/go-bank/pkg/logger"
	"github.com/skamranahmed/go-bank/pkg/testutils"
)

var (
	postgresTestContainer *testutils.PostgresTestContainer
	redisTestContainer    *testutils.RedisTestContainer
)

func TestMain(m *testing.M) {
	// init logger
	logger.Init()

	ctx := context.TODO()

	postgresTestContainer = testutils.NewPostgresTestContainer(ctx)
	redisTestContainer = testutils.NewRedisTestContainer(ctx)

	// run tests
	code := m.Run()

	// teardowns
	postgresTestContainer.TeardownFunc()
	redisTestContainer.TeardownFunc()

	// teardown
	os.Exit(code)
}
