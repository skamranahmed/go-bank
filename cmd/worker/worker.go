package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bsm/redislock"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/internal"
	userTasks "github.com/skamranahmed/go-bank/internal/user/tasks"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

func Start(queueName string, services *internal.Services) {
	ctx := context.TODO()
	redisConfig := config.GetRedisConfig()

	/*
		worker setup
	*/
	workerDone := make(chan struct{}) // workerDone channel signals when worker.start() exits
	worker := newWorker(queueName, redisConfig, services)
	go worker.start(ctx, workerDone)

	/*
		scheduler setup
	*/
	schedulerDone := make(chan struct{}) // schedulerDone channel signals when scheduler.start() exits
	scheduler := newScheduler(redisConfig)
	go scheduler.start(ctx, schedulerDone)

	<-workerDone // blocks until the worker stops
	logger.Info(ctx, "Worker stopped")

	<-schedulerDone // blocks until the scheduler stops
	logger.Info(ctx, "Scheduler stopped")
}

type worker struct {
	*asynq.Server
	TaskHandler *asynq.ServeMux
	Services    *internal.Services
}

func newWorker(queueName string, redisConfig config.RedisConfig, services *internal.Services) *worker {
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
		Services:    services,
	}
}

/*
start sets up task processors and runs the worker.
It blocks until an os signal to exit the program is received.
Once it receives a signal, it gracefully shuts down all active workers and other goroutines to process the tasks.
*/
func (w *worker) start(ctx context.Context, workerDone chan struct{}) {
	w.registerTaskProcessors()
	logger.Info(ctx, "Worker is starting")
	err := w.Run(w.TaskHandler)
	if err != nil {
		logger.Fatal(ctx, "Could not run worker server: %+v", err)
	}
	close(workerDone) // signal completion
}

func (w *worker) registerTaskProcessors() {
	// user tasks
	userTasks.RegisterTaskProcessors(w.TaskHandler, w.Services)
}

type scheduler struct {
	*asynq.Scheduler
	RedisConfig config.RedisConfig
}

func newScheduler(redisConfig config.RedisConfig) *scheduler {
	return &scheduler{
		Scheduler: asynq.NewScheduler(asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
			Password: redisConfig.Password,
			DB:       redisConfig.DbIndex,
		}, nil),
		RedisConfig: redisConfig,
	}
}

/*
start attempts to acquire an exclusive lock for the scheduler.
It blocks until the lock is acquired or an os signal to exit the program is received.
The function returns once the lock is held or shutdown is triggered.

NOTE: Asynq requires us to run only one scheduler instance at a time to prevent duplicate tasks. (https://git.new/v03JF3B)

However, a single scheduler becomes a SPOF (if it crashes, no tasks will be scheduled, until the instance is up and running again).

For HA, we use a redis distributed lock using Redlock:
  - Multiple schedulers can run in parallel.
  - Only the scheduler holding the lock will be allowed to enqueue tasks.
  - If the active scheduler crashes or loses the lock, another instance
    can acquire it and continue scheduling.
  - This ensures fault tolerance while preserving the "only one active scheduler"
    requirement of Asynq.
*/
func (s *scheduler) start(ctx context.Context, schedulerDone chan struct{}) {
	osSignalChannel := make(chan os.Signal, 1)
	signal.Notify(osSignalChannel, os.Interrupt, syscall.SIGTERM)

	// shutdownSignalChannel is to notify other go-routines to exit gracefully
	shutdownSignalChannel := make(chan struct{})

	go func() {
		/*
			- osSignalChannel receives OS signals such as SIGINT or SIGTERM
			- When such a signal is received, shutdownSignalChannel is closed
			  to notify all go-routines to exit gracefully
			- This separates OS signal handling from the actual shutdown mechanism
			  that other go-routines are listening to
		*/
		<-osSignalChannel
		close(shutdownSignalChannel)
	}()

	// register the scheduled tasks
	s.registerScheduledTasks()

	isSchedulerLockAcquired := s.acquireExclusiveLock(ctx, shutdownSignalChannel) // blocking operation
	if isSchedulerLockAcquired {
		logger.Info(ctx, "Scheduler lock acquired, starting scheduler...")

		go func() {
			// run the scheduler
			err := s.Run()
			if err != nil {
				logger.Fatal(ctx, "Could not run scheduler server: %+v", err)
			}
			close(schedulerDone) // signal completion
		}()
	} else {
		// even though we couldn't acquire the lock but we still need to close the schedulerDone channel
		// to signal completion so that server can be shutdown properly without blocking indefinitely
		close(schedulerDone)
	}
}

/*
acquireExclusiveLock repeatedly tries to acquire the exclusive scheduler lock,
pausing briefly between attempts. The pause is added to avoid overwhelming the redis server.

  - it blocks until the lock is acquired or until a shutdown signal is received
  - returns true once lock is successfully acquired
  - returns false if a shutdown signal is received and the lock wasn't acquired until that time
*/
func (s *scheduler) acquireExclusiveLock(ctx context.Context, shutdownSignalChannel chan struct{}) bool {
	var err error
	var schedulerLock *redislock.Lock

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", s.RedisConfig.Host, s.RedisConfig.Port),
		Password: s.RedisConfig.Password,
		DB:       s.RedisConfig.DbIndex,
	})

	redisLockClient := redislock.New(redisClient)
	schedulerLockKey := "asynq:scheduler:lock"

	// repeatedly try to acquire the lock
	for {
		select {
		// if we have received a shutdown signal, we don't need to try to acquire the lock anymore
		case <-shutdownSignalChannel:
			return false
		default:
			// NOTE: we are not passing a retry strategy to redisLockClient.Obtain()
			// This gives us full control over how we want to retry instead of relying on the library's implicit retries
			schedulerLock, err = redisLockClient.Obtain(ctx, schedulerLockKey, time.Minute, nil)
			if err == redislock.ErrNotObtained {
				// This means another scheduler instance must be holding the lock
				// we should pause for some time before retrying to obtain the lock
				logger.Info(ctx, "Scheduler lock currently held by another instance. Waiting to retry.")

				// The sleep time here should be thought carefully. If it is too large, there can be a gap between
				// another instance releasing the lock and us acquiring it, leading to delay in scheduling.
				// If it is too small, we will end up making frequent calls to redis. Hence, a balance is needed.
				// 5 secs is looking as a good sleep interval to begin with.
				time.Sleep(5 * time.Second)
				continue
			} else if err != nil {
				logger.Fatal(ctx, "Failed to acquire lock. Something might be wrong with redis, error: %+v", err)
			}

			/*
				=====================================================================================
					                At this point, the lock has been acquired.
					The go-routines defined below will never run if the lock hasn't been acquired.
				=====================================================================================
			*/

			/*
				The lock is periodically refreshed to extend its TTL, ensuring that this
				instance remains the scheduler and be responseilbe for enqueueing scheduled
				tasks for as long as it is healthy. This prevents unnecessary leadership handovers
				and avoids gaps where no scheduler is active. If the instance crashes or shuts down,
				the lock is released (or will naturally expire), allowing another instance to take over.

				The ticker triggers every 30s to refresh a 1-minute TTL. Each refresh extends
				the lock's expiry relative to the refresh time.

				Example:
				  00:00:00 --> lock acquired, TTL = 1m (expires at 00:01:00)
				  00:00:30 --> 1st refresh, TTL reset to 1m from now (expires at 00:01:30)
				  00:01:00 --> 2nd refresh, TTL reset to 1m from now (expires at 00:02:00)

				If Refresh returns ErrNotObtained, the lock is lost and we stop refreshing.
				Any other error is fatal because it indicates a problem with Redis.
			*/
			go func(schedulerLock *redislock.Lock) {
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()

				for range ticker.C {
					err := schedulerLock.Refresh(ctx, time.Minute, nil)
					if err == redislock.ErrNotObtained {
						// this will stop the ticker and we will no longer keep on refreshing the lock
						break
					} else if err != nil {
						logger.Fatal(ctx, "Failed to refresh scheduler lock. Something might be wrong with redis, error: %+v", err)
					}
				}
			}(schedulerLock)

			/*
				Listen for a shutdown signal (shutdownSignalChannel) and release the scheduler lock immediately.
				This ensures that if this instance is shutting down, the lock is freed right away,
				minimizing the time other scheduler instances have to wait to acquire the lock.
				Without this, other instances would have to wait until the lock's TTL expires.
			*/
			go func(shutdownSignalChannel chan struct{}) {
				<-shutdownSignalChannel
				logger.Info(ctx, "Shutdown signal received, releasing scheduler lock")
				schedulerLock.Release(ctx)
			}(shutdownSignalChannel)

			return true
		}
	}
}

func (s *scheduler) registerScheduledTasks() {
	// user tasks
	userTasks.RegisterScheduledTasks(s.Scheduler)
}
