package models

import (
	"time"
)

// Service представляет сервис для мониторинга
type Service struct {
	ID            int       `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	URL           string    `json:"url" db:"url"`
	CheckInterval int       `json:"check_interval" db:"check_interval"`
	Timeout       int       `json:"timeout" db:"timeout"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	LastStatus    string    `json:"last_status,omitempty"`
	LastCheck     time.Time `json:"last_check,omitempty"`
	Uptime        float64   `json:"uptime,omitempty"`
}

// HealthCheck представляет результат проверки здоровья сервиса
type HealthCheck struct {
	ID           int       `json:"id" db:"id"`
	ServiceID    int       `json:"service_id" db:"service_id"`
	Status       string    `json:"status" db:"status"`
	ResponseTime int       `json:"response_time" db:"response_time"`
	ErrorMessage string    `json:"error_message" db:"error_message"`
	CheckedAt    time.Time `json:"checked_at" db:"checked_at"`
}

// Alert представляет алерт о проблеме с сервисом
type Alert struct {
	ID         int       `json:"id" db:"id"`
	ServiceID  int       `json:"service_id" db:"service_id"`
	Message    string    `json:"message" db:"message"`
	Severity   string    `json:"severity" db:"severity"`
	IsResolved bool      `json:"is_resolved" db:"is_resolved"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at" db:"resolved_at"`
	Service    *Service  `json:"service,omitempty"`
}

// CreateServiceRequest запрос на создание сервиса
type CreateServiceRequest struct {
	Name          string `json:"name" binding:"required"`
	URL           string `json:"url" binding:"required"`
	CheckInterval int    `json:"check_interval"`
	Timeout       int    `json:"timeout"`
}

// UpdateServiceRequest запрос на обновление сервиса
type UpdateServiceRequest struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	CheckInterval int    `json:"check_interval"`
	Timeout       int    `json:"timeout"`
}

// DashboardStats статистика для дашборда
type DashboardStats struct {
	TotalServices    int     `json:"total_services"`
	HealthyServices  int     `json:"healthy_services"`
	UnhealthyServices int    `json:"unhealthy_services"`
	AverageUptime    float64 `json:"average_uptime"`
	ActiveAlerts     int     `json:"active_alerts"`
}

// ServiceStatus статус сервиса
const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
	StatusUnknown   = "unknown"
)

// AlertSeverity уровни важности алертов
const (
	SeverityInfo    = "info"
	SeverityWarning = "warning"
	SeverityError   = "error"
	SeverityCritical = "critical"
)
