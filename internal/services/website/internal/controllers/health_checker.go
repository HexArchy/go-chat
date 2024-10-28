package controllers

import (
	"context"
	"sync/atomic"

	"google.golang.org/grpc/health/grpc_health_v1"
)

type HealthChecker struct {
	grpc_health_v1.UnimplementedHealthServer
	status int32
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		status: int32(grpc_health_v1.HealthCheckResponse_SERVING),
	}
}

func (h *HealthChecker) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_ServingStatus(atomic.LoadInt32(&h.status)),
	}, nil
}

func (h *HealthChecker) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return stream.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_ServingStatus(atomic.LoadInt32(&h.status)),
	})
}

func (h *HealthChecker) SetStatus(status grpc_health_v1.HealthCheckResponse_ServingStatus) {
	atomic.StoreInt32(&h.status, int32(status))
}

func (h *HealthChecker) GetStatus() grpc_health_v1.HealthCheckResponse_ServingStatus {
	return grpc_health_v1.HealthCheckResponse_ServingStatus(atomic.LoadInt32(&h.status))
}
