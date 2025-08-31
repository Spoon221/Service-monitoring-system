package monitor

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"service-monitor/internal/database"
	"service-monitor/internal/logger"
	"service-monitor/internal/models"
)

type Service struct {
	db     *database.DB
	logger *logger.Logger
	ticker *time.Ticker
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewService(db *database.DB, logger *logger.Logger) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Service{
		db:     db,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Service) Start() {
	s.logger.Info("Запуск мониторинга сервисов")
	
	// Запускаем проверку каждые 30 секунд
	s.ticker = time.NewTicker(30 * time.Second)
	defer s.ticker.Stop()

	// Первая проверка сразу
	s.checkAllServices()

	for {
		select {
		case <-s.ticker.C:
			s.checkAllServices()
		case <-s.ctx.Done():
			s.logger.Info("Остановка мониторинга")
			return
		}
	}
}

func (s *Service) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *Service) checkAllServices() {
	services, err := s.getServices()
	if err != nil {
		s.logger.Error("Ошибка получения сервисов:", err)
		return
	}

	for _, service := range services {
		s.wg.Add(1)
		go func(service models.Service) {
			defer s.wg.Done()
			s.checkService(service)
		}(service)
	}
}

func (s *Service) checkService(service models.Service) {
	start := time.Now()
	
	client := &http.Client{
		Timeout: time.Duration(service.Timeout) * time.Second,
	}

	resp, err := client.Get(service.URL)
	responseTime := int(time.Since(start).Milliseconds())

	var status string
	var errorMessage string

	if err != nil {
		status = models.StatusUnhealthy
		errorMessage = err.Error()
		s.logger.Error("Сервис", service.Name, "недоступен:", err)
	} else {
		defer resp.Body.Close()
		
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			status = models.StatusHealthy
		} else {
			status = models.StatusUnhealthy
			errorMessage = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
	}

	// Сохраняем результат проверки
	check := models.HealthCheck{
		ServiceID:    service.ID,
		Status:       status,
		ResponseTime: responseTime,
		ErrorMessage: errorMessage,
		CheckedAt:    time.Now(),
	}

	if err := s.saveHealthCheck(check); err != nil {
		s.logger.Error("Ошибка сохранения проверки:", err)
		return
	}

	// Проверяем необходимость создания алерта
	if status == models.StatusUnhealthy {
		s.checkAndCreateAlert(service, errorMessage)
	} else {
		// Если сервис восстановился, разрешаем алерты
		s.resolveAlerts(service.ID)
	}
}

func (s *Service) getServices() ([]models.Service, error) {
	query := `SELECT id, name, url, check_interval, timeout, created_at, updated_at FROM services`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var service models.Service
		err := rows.Scan(
			&service.ID,
			&service.Name,
			&service.URL,
			&service.CheckInterval,
			&service.Timeout,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

func (s *Service) saveHealthCheck(check models.HealthCheck) error {
	query := `
		INSERT INTO health_checks (service_id, status, response_time, error_message, checked_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	
	_, err := s.db.Exec(query, check.ServiceID, check.Status, check.ResponseTime, check.ErrorMessage, check.CheckedAt)
	return err
}

func (s *Service) checkAndCreateAlert(service models.Service, errorMessage string) {
	// Проверяем, есть ли уже активный алерт для этого сервиса
	query := `SELECT COUNT(*) FROM alerts WHERE service_id = $1 AND is_resolved = false`
	
	var count int
	err := s.db.QueryRow(query, service.ID).Scan(&count)
	if err != nil {
		s.logger.Error("Ошибка проверки алертов:", err)
		return
	}

	// Если алерта нет, создаем новый
	if count == 0 {
		alert := models.Alert{
			ServiceID:  service.ID,
			Message:    fmt.Sprintf("Сервис %s недоступен: %s", service.Name, errorMessage),
			Severity:   models.SeverityError,
			IsResolved: false,
			CreatedAt:  time.Now(),
		}

		insertQuery := `
			INSERT INTO alerts (service_id, message, severity, is_resolved, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`
		
		_, err := s.db.Exec(insertQuery, alert.ServiceID, alert.Message, alert.Severity, alert.IsResolved, alert.CreatedAt)
		if err != nil {
			s.logger.Error("Ошибка создания алерта:", err)
		} else {
			s.logger.Info("Создан алерт для сервиса:", service.Name)
		}
	}
}

func (s *Service) resolveAlerts(serviceID int) error {
	query := `
		UPDATE alerts 
		SET is_resolved = true, resolved_at = CURRENT_TIMESTAMP 
		WHERE service_id = $1 AND is_resolved = false
	`
	
	_, err := s.db.Exec(query, serviceID)
	if err != nil {
		s.logger.Error("Ошибка разрешения алертов:", err)
	}
	
	return err
}
