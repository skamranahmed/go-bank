package internal

import (
	accountRepository "github.com/skamranahmed/go-bank/internal/account/repository"
	accountService "github.com/skamranahmed/go-bank/internal/account/service"
	authenticationService "github.com/skamranahmed/go-bank/internal/authentication/service"
	healthzService "github.com/skamranahmed/go-bank/internal/healthz/service"
	transferService "github.com/skamranahmed/go-bank/internal/transfer/service"
	userRepository "github.com/skamranahmed/go-bank/internal/user/repository"
	userService "github.com/skamranahmed/go-bank/internal/user/service"
	"github.com/skamranahmed/go-bank/pkg/cache"
	tasksHelper "github.com/skamranahmed/go-bank/pkg/tasks"
	"github.com/uptrace/bun"
)

type Services struct {
	AccountService        accountService.AccountService
	AuthenticationService authenticationService.AuthenticationService
	HealthzService        healthzService.HealthzService
	TaskEnqueuer          tasksHelper.TaskEnqueuer
	TransferService       transferService.TransferService
	UserService           userService.UserService
}

func BootstrapServices(db *bun.DB, cacheClient cache.CacheClient, taskEnqueuer tasksHelper.TaskEnqueuer) (*Services, error) {
	// healthz service
	healthzService := healthzService.NewHealthzService(db, cacheClient)

	// user service
	userRepository := userRepository.NewUserRepository(db)
	userService := userService.NewUserService(db, userRepository)

	// authentication service
	authenticationService := authenticationService.NewAuthenticationService(db, cacheClient)

	// account service
	accountRepository := accountRepository.NewAccountRepository(db)
	accountService := accountService.NewAccountService(db, accountRepository)

	// transfer service
	transferService := transferService.NewTransferService(db, accountService)

	return &Services{
		AccountService:        accountService,
		AuthenticationService: authenticationService,
		HealthzService:        healthzService,
		TaskEnqueuer:          taskEnqueuer,
		TransferService:       transferService,
		UserService:           userService,
	}, nil
}
