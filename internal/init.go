package internal

import (
	accountRepository "github.com/skamranahmed/go-bank/internal/account/repository"
	accountService "github.com/skamranahmed/go-bank/internal/account/service"
	authenticationService "github.com/skamranahmed/go-bank/internal/authentication/service"
	healthzService "github.com/skamranahmed/go-bank/internal/healthz/service"
	userRepository "github.com/skamranahmed/go-bank/internal/user/repository"
	userService "github.com/skamranahmed/go-bank/internal/user/service"
	"github.com/skamranahmed/go-bank/pkg/cache"
	"github.com/uptrace/bun"
)

type Services struct {
	HealthzService        healthzService.HealthzService
	AuthenticationService authenticationService.AuthenticationService
	UserService           userService.UserService
	AccountService        accountService.AccountService
}

func BootstrapServices(db *bun.DB, cacheClient cache.CacheClient) (*Services, error) {
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

	return &Services{
		HealthzService:        healthzService,
		AuthenticationService: authenticationService,
		UserService:           userService,
		AccountService:        accountService,
	}, nil
}
