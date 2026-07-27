package service_api_impl

import (
	"testing"
	"time"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/service_manager/verv_services"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	testServiceName = "test-service"
)

// Test that ServiceBaseInfo is properly converted with all enriched fields.
func TestToServiceBaseInfoWithEnrichedFields(t *testing.T) {
	now := time.Now()
	input := domain.ServiceBaseInfo{
		Name:           testServiceName,
		LastDeployedAt: &now,
		ImageName:      "docker.io/test-service:v1.0",
		Status:         verv_services.StatusRunning,
		Env:            "production",
	}

	result := toServiceBaseInfo(input)

	if result.GetName() != testServiceName {
		t.Errorf("expected Name 'test-service', got %q", result.GetName())
	}

	if result.GetImageName() != "docker.io/test-service:v1.0" {
		t.Errorf("expected ImageName 'docker.io/test-service:v1.0', got %q", result.GetImageName())
	}

	if result.GetStatus() != "running" {
		t.Errorf("expected Status 'running', got %q", result.GetStatus())
	}

	if result.GetEnv() != "production" {
		t.Errorf("expected Env 'production', got %q", result.GetEnv())
	}

	if result.GetLastDeployedAt() == nil {
		t.Error("expected LastDeployedAt to be set, got nil")
	}

	expectedTime := timestamppb.New(now)
	if result.GetLastDeployedAt().GetSeconds() != expectedTime.GetSeconds() {
		t.Errorf(
			"expected LastDeployedAt seconds %d, got %d",
			expectedTime.GetSeconds(),
			result.GetLastDeployedAt().GetSeconds(),
		)
	}
}

// Test that ServiceBaseInfo conversion handles empty fields.
func TestToServiceBaseInfoWithEmptyFields(t *testing.T) {
	input := domain.ServiceBaseInfo{
		Name:      "test-service",
		ImageName: "",
		Status:    "",
		Env:       "",
	}

	result := toServiceBaseInfo(input)

	if result.GetName() != "test-service" {
		t.Errorf("expected Name 'test-service', got %q", result.GetName())
	}

	if result.GetImageName() != "" {
		t.Errorf("expected empty ImageName, got %q", result.GetImageName())
	}

	if result.GetStatus() != "" {
		t.Errorf("expected empty Status, got %q", result.GetStatus())
	}

	if result.GetEnv() != "" {
		t.Errorf("expected empty Env, got %q", result.GetEnv())
	}

	if result.GetLastDeployedAt() != nil {
		t.Error("expected LastDeployedAt to be nil, got set")
	}
}

// Test conversion of list.
func TestToServiceBaseInfoList(t *testing.T) {
	now := time.Now()
	input := []domain.ServiceBaseInfo{
		{
			Name:           "service1",
			LastDeployedAt: &now,
			ImageName:      "image1",
			Status:         "running",
			Env:            "prod",
		},
		{
			Name:      "service2",
			ImageName: "image2",
			Status:    "stopped",
		},
	}

	result := toServiceBaseInfoList(input)

	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}

	if result[0].GetName() != "service1" {
		t.Errorf("expected first service name 'service1', got %q", result[0].GetName())
	}

	if result[0].GetImageName() != "image1" {
		t.Errorf("expected first service image 'image1', got %q", result[0].GetImageName())
	}

	if result[1].GetName() != "service2" {
		t.Errorf("expected second service name 'service2', got %q", result[1].GetName())
	}

	if result[1].GetStatus() != "stopped" {
		t.Errorf("expected second service status 'stopped', got %q", result[1].GetStatus())
	}
}
