package controllers

import (
	"context"

	"google.golang.org/grpc/health/grpc_health_v1"
)

// HealthChecker implements custom grpc_health_v1.HealthServer.
type HealthChecker struct {
	grpc_health_v1.UnimplementedHealthServer
	status grpc_health_v1.HealthCheckResponse_ServingStatus
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		status: grpc_health_v1.HealthCheckResponse_SERVING,
	}
}

func (h *HealthChecker) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: h.status,
	}, nil
}

func (h *HealthChecker) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return stream.Send(&grpc_health_v1.HealthCheckResponse{
		Status: h.status,
	})
}

func (h *HealthChecker) SetStatus(status grpc_health_v1.HealthCheckResponse_ServingStatus) {
	h.status = status
}
