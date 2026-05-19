package service_api_impl

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.vervstack.ru/Velez/internal/domain"
)

// Test that ServiceBaseInfo is properly converted with all enriched fields
func TestToServiceBaseInfoWithEnrichedFields(t *testing.T) {
	now := time.Now()
	input := domain.ServiceBaseInfo{
		Name:           "test-service",
		LastDeployedAt: &now,
		ImageName:      "docker.io/test-service:v1.0",
		Status:         "running",
		Env:            "production",
	}

	result := toServiceBaseInfo(input)

	if result.Name != "test-service" {
		t.Errorf("expected Name 'test-service', got %q", result.Name)
	}

	if result.ImageName != "docker.io/test-service:v1.0" {
		t.Errorf("expected ImageName 'docker.io/test-service:v1.0', got %q", result.ImageName)
	}

	if result.Status != "running" {
		t.Errorf("expected Status 'running', got %q", result.Status)
	}

	if result.Env != "production" {
		t.Errorf("expected Env 'production', got %q", result.Env)
	}

	if result.LastDeployedAt == nil {
		t.Error("expected LastDeployedAt to be set, got nil")
	}

	expectedTime := timestamppb.New(now)
	if result.LastDeployedAt.Seconds != expectedTime.Seconds {
		t.Errorf("expected LastDeployedAt seconds %d, got %d", expectedTime.Seconds, result.LastDeployedAt.Seconds)
	}
}

// Test that ServiceBaseInfo conversion handles empty fields
func TestToServiceBaseInfoWithEmptyFields(t *testing.T) {
	input := domain.ServiceBaseInfo{
		Name:      "test-service",
		ImageName: "",
		Status:    "",
		Env:       "",
	}

	result := toServiceBaseInfo(input)

	if result.Name != "test-service" {
		t.Errorf("expected Name 'test-service', got %q", result.Name)
	}

	if result.ImageName != "" {
		t.Errorf("expected empty ImageName, got %q", result.ImageName)
	}

	if result.Status != "" {
		t.Errorf("expected empty Status, got %q", result.Status)
	}

	if result.Env != "" {
		t.Errorf("expected empty Env, got %q", result.Env)
	}

	if result.LastDeployedAt != nil {
		t.Error("expected LastDeployedAt to be nil, got set")
	}
}

// Test conversion of list
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

	if result[0].Name != "service1" {
		t.Errorf("expected first service name 'service1', got %q", result[0].Name)
	}

	if result[0].ImageName != "image1" {
		t.Errorf("expected first service image 'image1', got %q", result[0].ImageName)
	}

	if result[1].Name != "service2" {
		t.Errorf("expected second service name 'service2', got %q", result[1].Name)
	}

	if result[1].Status != "stopped" {
		t.Errorf("expected second service status 'stopped', got %q", result[1].Status)
	}
}
