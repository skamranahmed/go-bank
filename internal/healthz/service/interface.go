package service

type HealthzService interface {
	DbPing() error
}
