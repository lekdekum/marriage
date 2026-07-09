package services

import "context"

type HealthService struct{}

type HealthStatus struct {
	Status string `json:"status"`
}

func NewHealthService() HealthService {
	return HealthService{}
}

func (service HealthService) Check(ctx context.Context) HealthStatus {
	return HealthStatus{
		Status: "ok",
	}
}
