package verv_services

import (
	"context"
	"testing"
	"time"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

const (
	testServiceName    = "my-service"
	testStatusRunning  = "running"
	testStatusDegraded = "degraded"
	testStatusStopped  = "stopped"
)

// Mock container service for testing.
type testContainerService struct {
	listSmerdsFunc func(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error)
}

func (m *testContainerService) ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
	if m.listSmerdsFunc != nil {
		return m.listSmerdsFunc(ctx, req)
	}

	return &velez_api.ListSmerds_Response{}, nil
}

func (m *testContainerService) DropSmerds(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error) {
	return nil, nil
}

func (m *testContainerService) InspectSmerd(ctx context.Context, contId string) (*velez_api.Smerd, error) {
	return nil, nil
}

func (m *testContainerService) ConnectToNetwork(ctx context.Context, req domain.Connection) error {
	return nil
}

func (m *testContainerService) DisconnectFromNetwork(ctx context.Context, req domain.Connection) error {
	return nil
}

// Mock storage service for testing.
type testStorageService struct {
	listFunc func(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error)
}

func (m *testStorageService) List(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, req)
	}

	return domain.ServiceList{}, nil
}

func (m *testStorageService) GetByName(ctx context.Context, name string) (domain.Service, error) {
	return domain.Service{}, nil
}

func (m *testStorageService) UpsertService(ctx context.Context, name string) error {
	return nil
}

func (m *testStorageService) Delete(ctx context.Context, name string) error {
	return nil
}

// Test that services are enriched with smerd data.
func TestEnrichServiceWithSmerdData(t *testing.T) {
	baseInfo := domain.ServiceBaseInfo{
		Name: testServiceName,
	}

	mockStorage := &testStorageService{
		listFunc: func(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error) {
			return domain.ServiceList{
				Total:    1,
				Services: []domain.ServiceBaseInfo{baseInfo},
			}, nil
		},
	}

	mockContainer := &testContainerService{
		listSmerdsFunc: func(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
			return &velez_api.ListSmerds_Response{
				Smerds: []*velez_api.Smerd{
					{
						Name:      testServiceName,
						ImageName: "docker.io/my-service:latest",
						Status:    velez_api.Smerd_running,
						Labels: map[string]string{
							"env":        "production",
							"velez.repo": "https://github.com/vervstack/my-service",
						},
					},
				},
			}, nil
		},
	}

	service := &VervService{
		servicesStorage:  mockStorage,
		containerService: mockContainer,
	}

	result, err := service.List(context.Background(), domain.ListServicesReq{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(result.Services))
	}

	enrichedService := result.Services[0]
	if enrichedService.Name != testServiceName {
		t.Errorf("expected name %q, got %q", testServiceName, enrichedService.Name)
	}

	if enrichedService.ImageName != "docker.io/my-service:latest" {
		t.Errorf("expected image 'docker.io/my-service:latest', got %q", enrichedService.ImageName)
	}

	if enrichedService.Status != testStatusRunning {
		t.Errorf("expected status %q, got %q", testStatusRunning, enrichedService.Status)
	}

	if enrichedService.Env != "production" {
		t.Errorf("expected env 'production', got %q", enrichedService.Env)
	}

	if enrichedService.Repo != "https://github.com/vervstack/my-service" {
		t.Errorf("expected repo 'https://github.com/vervstack/my-service', got %q", enrichedService.Repo)
	}
}

// Test that services without smerds are left with empty fields.
func TestServiceWithoutSmerd(t *testing.T) {
	baseInfo := domain.ServiceBaseInfo{
		Name: testServiceName,
	}

	mockStorage := &testStorageService{
		listFunc: func(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error) {
			return domain.ServiceList{
				Total:    1,
				Services: []domain.ServiceBaseInfo{baseInfo},
			}, nil
		},
	}

	mockContainer := &testContainerService{
		listSmerdsFunc: func(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
			return &velez_api.ListSmerds_Response{
				Smerds: []*velez_api.Smerd{},
			}, nil
		},
	}

	service := &VervService{
		servicesStorage:  mockStorage,
		containerService: mockContainer,
	}

	result, err := service.List(context.Background(), domain.ListServicesReq{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(result.Services))
	}

	enrichedService := result.Services[0]
	if enrichedService.Name != testServiceName {
		t.Errorf("expected name %q, got %q", testServiceName, enrichedService.Name)
	}

	if enrichedService.ImageName != "" {
		t.Errorf("expected empty image, got %q", enrichedService.ImageName)
	}

	if enrichedService.Status != "" {
		t.Errorf("expected empty status, got %q", enrichedService.Status)
	}

	if enrichedService.Env != "" {
		t.Errorf("expected empty env, got %q", enrichedService.Env)
	}

	if enrichedService.Repo != "" {
		t.Errorf("expected empty repo, got %q", enrichedService.Repo)
	}
}

// Test smerd status mapping.
func TestMapSmerdStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   velez_api.Smerd_Status
		expected string
	}{
		{"running", velez_api.Smerd_running, testStatusRunning},
		{"paused", velez_api.Smerd_paused, testStatusDegraded},
		{"exited", velez_api.Smerd_exited, testStatusStopped},
		{"dead", velez_api.Smerd_dead, testStatusStopped},
		{"unknown", velez_api.Smerd_unknown, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapSmerdStatus(tt.status)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Test that services with LastDeployedAt preserve that data.
func TestServicePreservesLastDeployedAt(t *testing.T) {
	now := time.Now()
	baseInfo := domain.ServiceBaseInfo{
		Name:           testServiceName,
		LastDeployedAt: &now,
	}

	mockStorage := &testStorageService{
		listFunc: func(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error) {
			return domain.ServiceList{
				Total:    1,
				Services: []domain.ServiceBaseInfo{baseInfo},
			}, nil
		},
	}

	mockContainer := &testContainerService{
		listSmerdsFunc: func(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error) {
			return &velez_api.ListSmerds_Response{
				Smerds: []*velez_api.Smerd{
					{
						Name:      testServiceName,
						ImageName: "my-image:v1",
						Status:    velez_api.Smerd_running,
					},
				},
			}, nil
		},
	}

	service := &VervService{
		servicesStorage:  mockStorage,
		containerService: mockContainer,
	}

	result, err := service.List(context.Background(), domain.ListServicesReq{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	enrichedService := result.Services[0]
	if enrichedService.LastDeployedAt == nil {
		t.Error("expected LastDeployedAt to be set, got nil")
	} else if *enrichedService.LastDeployedAt != now {
		t.Errorf("expected LastDeployedAt %v, got %v", now, *enrichedService.LastDeployedAt)
	}
}
