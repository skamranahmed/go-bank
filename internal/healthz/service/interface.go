package service

type HealthzService interface {
	DbPing() error
	CachePing() error
}
