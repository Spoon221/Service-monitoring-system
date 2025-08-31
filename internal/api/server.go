package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"service-monitor/internal/config"
	"service-monitor/internal/database"
	"service-monitor/internal/logger"
	"service-monitor/internal/monitor"
	"service-monitor/internal/models"
)

type Server struct {
	config         *config.Config
	db             *database.DB
	monitorService *monitor.Service
	logger         *logger.Logger
	upgrader       websocket.Upgrader
	clients        map[*websocket.Conn]bool
}

func NewServer(cfg *config.Config, db *database.DB, monitorService *monitor.Service, logger *logger.Logger) *Server {
	return &Server{
		config:         cfg,
		db:             db,
		monitorService: monitorService,
		logger:         logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true 
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) Run() error {
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Статические файлы
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	// Главная страница
	router.GET("/", s.handleDashboard)

	// API endpoints
	api := router.Group("/api/v1")
	{
		// Сервисы
		api.GET("/services", s.getServices)
		api.POST("/services", s.createService)
		api.GET("/services/:id", s.getService)
		api.PUT("/services/:id", s.updateService)
		api.DELETE("/services/:id", s.deleteService)
		
		// Проверки здоровья
		api.GET("/services/:id/checks", s.getServiceChecks)
		
		// Алерты
		api.GET("/alerts", s.getAlerts)
		api.PUT("/alerts/:id/resolve", s.resolveAlert)
		
		// Статистика
		api.GET("/stats", s.getStats)
		
		// WebSocket для real-time обновлений
		api.GET("/ws", s.handleWebSocket)
	}

	return router.Run(":" + s.config.Port)
}

func (s *Server) handleDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Система мониторинга сервисов",
	})
}

func (s *Server) getServices(c *gin.Context) {
	query := `
		SELECT s.*, 
		       hc.status as last_status,
		       hc.checked_at as last_check,
		       COALESCE(
		           (SELECT COUNT(*) * 100.0 / NULLIF(COUNT(*) OVER(), 0)
		            FROM health_checks hc2 
		            WHERE hc2.service_id = s.id 
		            AND hc2.status = 'healthy'
		            AND hc2.checked_at >= NOW() - INTERVAL '24 hours'), 0
		       ) as uptime
		FROM services s
		LEFT JOIN LATERAL (
			SELECT status, checked_at 
			FROM health_checks 
			WHERE service_id = s.id 
			ORDER BY checked_at DESC 
			LIMIT 1
		) hc ON true
		ORDER BY s.created_at DESC
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var service models.Service
		var lastStatus, lastCheck sql.NullString
		var uptime sql.NullFloat64
		
		err := rows.Scan(
			&service.ID,
			&service.Name,
			&service.URL,
			&service.CheckInterval,
			&service.Timeout,
			&service.CreatedAt,
			&service.UpdatedAt,
			&lastStatus,
			&lastCheck,
			&uptime,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if lastStatus.Valid {
			service.LastStatus = lastStatus.String
		}
		if lastCheck.Valid {
			if t, err := time.Parse(time.RFC3339, lastCheck.String); err == nil {
				service.LastCheck = t
			}
		}
		if uptime.Valid {
			service.Uptime = uptime.Float64
		}

		services = append(services, service)
	}

	c.JSON(http.StatusOK, services)
}

func (s *Server) createService(c *gin.Context) {
	var req models.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Устанавливаем значения по умолчанию
	if req.CheckInterval == 0 {
		req.CheckInterval = 30
	}
	if req.Timeout == 0 {
		req.Timeout = 10
	}

	query := `
		INSERT INTO services (name, url, check_interval, timeout)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	
	var service models.Service
	err := s.db.QueryRow(query, req.Name, req.URL, req.CheckInterval, req.Timeout).
		Scan(&service.ID, &service.CreatedAt, &service.UpdatedAt)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	service.Name = req.Name
	service.URL = req.URL
	service.CheckInterval = req.CheckInterval
	service.Timeout = req.Timeout

	c.JSON(http.StatusCreated, service)
}

func (s *Server) getService(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	query := `SELECT id, name, url, check_interval, timeout, created_at, updated_at FROM services WHERE id = $1`
	
	var service models.Service
	err = s.db.QueryRow(query, id).Scan(
		&service.ID,
		&service.Name,
		&service.URL,
		&service.CheckInterval,
		&service.Timeout,
		&service.CreatedAt,
		&service.UpdatedAt,
	)
	
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сервис не найден"})
		return
	}

	c.JSON(http.StatusOK, service)
}

func (s *Server) updateService(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var req models.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		UPDATE services 
		SET name = COALESCE($2, name),
		    url = COALESCE($3, url),
		    check_interval = COALESCE($4, check_interval),
		    timeout = COALESCE($5, timeout),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, name, url, check_interval, timeout, created_at, updated_at
	`
	
	var service models.Service
	err = s.db.QueryRow(query, id, req.Name, req.URL, req.CheckInterval, req.Timeout).Scan(
		&service.ID,
		&service.Name,
		&service.URL,
		&service.CheckInterval,
		&service.Timeout,
		&service.CreatedAt,
		&service.UpdatedAt,
	)
	
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сервис не найден"})
		return
	}

	c.JSON(http.StatusOK, service)
}

func (s *Server) deleteService(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	query := `DELETE FROM services WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сервис не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Сервис удален"})
}

func (s *Server) getServiceChecks(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	query := `
		SELECT id, service_id, status, response_time, error_message, checked_at
		FROM health_checks
		WHERE service_id = $1
		ORDER BY checked_at DESC
		LIMIT $2
	`
	
	rows, err := s.db.Query(query, id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var checks []models.HealthCheck
	for rows.Next() {
		var check models.HealthCheck
		err := rows.Scan(
			&check.ID,
			&check.ServiceID,
			&check.Status,
			&check.ResponseTime,
			&check.ErrorMessage,
			&check.CheckedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		checks = append(checks, check)
	}

	c.JSON(http.StatusOK, checks)
}

func (s *Server) getAlerts(c *gin.Context) {
	query := `
		SELECT a.id, a.service_id, a.message, a.severity, a.is_resolved, a.created_at, a.resolved_at,
		       s.name as service_name
		FROM alerts a
		LEFT JOIN services s ON a.service_id = s.id
		ORDER BY a.created_at DESC
		LIMIT 100
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var alert models.Alert
		var serviceName sql.NullString
		var resolvedAt sql.NullString
		
		err := rows.Scan(
			&alert.ID,
			&alert.ServiceID,
			&alert.Message,
			&alert.Severity,
			&alert.IsResolved,
			&alert.CreatedAt,
			&resolvedAt,
			&serviceName,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if serviceName.Valid {
			alert.Service = &models.Service{Name: serviceName.String}
		}
		if resolvedAt.Valid {
			if t, err := time.Parse(time.RFC3339, resolvedAt.String); err == nil {
				alert.ResolvedAt = &t
			}
		}

		alerts = append(alerts, alert)
	}

	c.JSON(http.StatusOK, alerts)
}

func (s *Server) resolveAlert(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	query := `
		UPDATE alerts 
		SET is_resolved = true, resolved_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	
	result, err := s.db.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Алерт не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Алерт разрешен"})
}

func (s *Server) getStats(c *gin.Context) {
	// Общая статистика
	statsQuery := `
		SELECT 
			COUNT(*) as total_services,
			COUNT(CASE WHEN hc.status = 'healthy' THEN 1 END) as healthy_services,
			COUNT(CASE WHEN hc.status = 'unhealthy' THEN 1 END) as unhealthy_services,
			COUNT(CASE WHEN a.is_resolved = false THEN 1 END) as active_alerts
		FROM services s
		LEFT JOIN LATERAL (
			SELECT status 
			FROM health_checks 
			WHERE service_id = s.id 
			ORDER BY checked_at DESC 
			LIMIT 1
		) hc ON true
		LEFT JOIN alerts a ON a.service_id = s.id AND a.is_resolved = false
	`
	
	var stats models.DashboardStats
	err := s.db.QueryRow(statsQuery).Scan(
		&stats.TotalServices,
		&stats.HealthyServices,
		&stats.UnhealthyServices,
		&stats.ActiveAlerts,
	)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Средний uptime за последние 24 часа
	uptimeQuery := `
		SELECT COALESCE(AVG(uptime), 0)
		FROM (
			SELECT 
				s.id,
				(COUNT(CASE WHEN hc.status = 'healthy' THEN 1 END) * 100.0 / COUNT(*)) as uptime
			FROM services s
			LEFT JOIN health_checks hc ON hc.service_id = s.id 
				AND hc.checked_at >= NOW() - INTERVAL '24 hours'
			GROUP BY s.id
		) uptimes
	`
	
	err = s.db.QueryRow(uptimeQuery).Scan(&stats.AverageUptime)
	if err != nil {
		stats.AverageUptime = 0
	}

	c.JSON(http.StatusOK, stats)
}

func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	s.clients[conn] = true
	defer delete(s.clients, conn)

	// Отправляем приветственное сообщение
	welcome := map[string]interface{}{
		"type": "welcome",
		"message": "Подключение к системе мониторинга установлено",
		"timestamp": time.Now(),
	}
	
	if err := conn.WriteJSON(welcome); err != nil {
		s.logger.Error("WebSocket write error:", err)
		return
	}

	// Ожидаем сообщения от клиента (для ping/pong)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// broadcastToClients отправляет сообщение всем подключенным WebSocket клиентам
func (s *Server) broadcastToClients(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		s.logger.Error("JSON marshal error:", err)
		return
	}

	for client := range s.clients {
		if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
			s.logger.Error("WebSocket broadcast error:", err)
			delete(s.clients, client)
			client.Close()
		}
	}
}
