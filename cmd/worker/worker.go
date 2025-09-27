package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/internal"
	userTasks "github.com/skamranahmed/go-bank/internal/user/tasks"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

func Start(queueName string, services *internal.Services) {
	ctx := context.TODO()
	redisConfig := config.GetRedisConfig()

	workerDone := make(chan struct{})    // workerDone channel signals when worker.start() exits
	schedulerDone := make(chan struct{}) // schedulerDone channel signals when scheduler.start() exits

	// worker setup
	worker := newWorker(queueName, redisConfig)
	worker.registerTaskProcessors(services)
	worker.start(ctx, workerDone)
	logger.Info(ctx, "Worker is running")

	// scheduler setup
	scheduler := newScheduler(redisConfig)
	scheduler.registerScheduledTasks()
	scheduler.start(ctx, schedulerDone)
	logger.Info(ctx, "Scheduler is running")

	<-workerDone // block until worker stops
	logger.Info(ctx, "Worker stopped")

	<-schedulerDone // block until scheduler stops
	logger.Info(ctx, "Scheduler stopped")
}

type worker struct {
	*asynq.Server
	TaskHandler *asynq.ServeMux
}

func newWorker(queueName string, redisConfig config.RedisConfig) *worker {
	return &worker{
		Server: asynq.NewServer(
			asynq.RedisClientOpt{
				Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
				Password: redisConfig.Password,
				DB:       redisConfig.DbIndex,
			},
			asynq.Config{
				Concurrency: 1,
				Queues: map[string]int{
					queueName: 1,
				},
			},
		),
		TaskHandler: asynq.NewServeMux(),
	}
}

func (w *worker) start(ctx context.Context, workerDone chan struct{}) {
	go func() {
		err := w.Run(w.TaskHandler)
		if err != nil {
			logger.Fatal(ctx, "Could not run worker server: %+v", err)
		}
		close(workerDone) // signal completion
	}()
}

func (w *worker) registerTaskProcessors(services *internal.Services) {
	// user tasks
	userTasks.RegisterTaskProcessors(w.TaskHandler, services)
}

type scheduler struct {
	*asynq.Scheduler
}

func newScheduler(redisConfig config.RedisConfig) *scheduler {
	return &scheduler{
		Scheduler: asynq.NewScheduler(asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
			Password: redisConfig.Password,
			DB:       redisConfig.DbIndex,
		}, nil),
	}
}

/*
NOTE: Asynq requires us to run only one scheduler instance at a time to prevent duplicate tasks. (https://git.new/v03JF3B)

However, a single scheduler becomes a SPOF (if it crashes, no tasks will be scheduled, until the instance is up and running again).

TODO: For HA, I am thinking to use a distributed lock (e.g: Redlock):
  - Multiple schedulers can run in parallel.
  - Only the scheduler holding the lock will be allowed to enqueue tasks.
  - If the active scheduler crashes or loses the lock, another instance
    can acquire it and continue scheduling.
    This ensures fault tolerance while preserving the "only one active scheduler"
    requirement of Asynq.
*/
func (s *scheduler) start(ctx context.Context, schedulerDone chan struct{}) {
	go func() {
		err := s.Run()
		if err != nil {
			logger.Fatal(ctx, "Could not run scheduler server: %+v", err)
		}
		close(schedulerDone) // signal completion
	}()
}

func (s *scheduler) registerScheduledTasks() {
	// user tasks
	userTasks.RegisterScheduledTasks(s.Scheduler)
}
