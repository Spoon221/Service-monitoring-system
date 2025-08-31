package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"service-monitor/internal/config"
	"service-monitor/internal/database"
	"service-monitor/internal/logger"
	"service-monitor/internal/monitor"
	"service-monitor/internal/models"
)

// Mock для базы данных
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Query(query string, args ...interface{}) (*database.DB, error) {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*database.DB), mockArgs.Error(1)
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *database.DB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*database.DB)
}

func (m *MockDB) Exec(query string, args ...interface{}) (interface{}, error) {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0), mockArgs.Error(1)
}

func (m *MockDB) Close() error {
	mockArgs := m.Called()
	return mockArgs.Error(0)
}

// Mock для мониторинга
type MockMonitor struct {
	mock.Mock
}

func (m *MockMonitor) Start() {
	m.Called()
}

func (m *MockMonitor) Stop() {
	m.Called()
}

func setupTestServer() (*gin.Engine, *MockDB, *MockMonitor) {
	gin.SetMode(gin.TestMode)
	
	mockDB := &MockDB{}
	mockMonitor := &MockMonitor{}
	
	cfg := &config.Config{
		Port: "8080",
	}
	
	logger := logger.New()
	
	server := NewServer(cfg, mockDB, mockMonitor, logger)
	
	router := gin.New()
	router.Use(gin.Recovery())
	
	// Настройка маршрутов
	api := router.Group("/api/v1")
	{
		api.GET("/services", server.getServices)
		api.POST("/services", server.createService)
		api.GET("/services/:id", server.getService)
		api.PUT("/services/:id", server.updateService)
		api.DELETE("/services/:id", server.deleteService)
		api.GET("/services/:id/checks", server.getServiceChecks)
		api.GET("/alerts", server.getAlerts)
		api.PUT("/alerts/:id/resolve", server.resolveAlert)
		api.GET("/stats", server.getStats)
	}
	
	return router, mockDB, mockMonitor
}

func TestGetServices(t *testing.T) {
	router, mockDB, _ := setupTestServer()
	
	// Настройка mock
	mockDB.On("Query", mock.AnythingOfType("string")).Return(&database.DB{}, nil)
	
	// Создание запроса
	req, _ := http.NewRequest("GET", "/api/v1/services", nil)
	w := httptest.NewRecorder()
	
	// Выполнение запроса
	router.ServeHTTP(w, req)
	
	// Проверки
	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestCreateService(t *testing.T) {
	router, mockDB, _ := setupTestServer()
	
	// Данные для создания сервиса
	serviceData := models.CreateServiceRequest{
		Name:          "Test Service",
		URL:           "https://example.com",
		CheckInterval: 30,
		Timeout:       10,
	}
	
	jsonData, _ := json.Marshal(serviceData)
	
	// Настройка mock
	mockDB.On("QueryRow", mock.AnythingOfType("string"), mock.Anything).Return(&database.DB{})
	
	// Создание запроса
	req, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Выполнение запроса
	router.ServeHTTP(w, req)
	
	// Проверки
	assert.Equal(t, http.StatusCreated, w.Code)
	mockDB.AssertExpectations(t)
}

func TestCreateServiceInvalidData(t *testing.T) {
	router, _, _ := setupTestServer()
	
	// Неверные данные
	invalidData := `{"name": "", "url": "invalid-url"}`
	
	// Создание запроса
	req, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBufferString(invalidData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Выполнение запроса
	router.ServeHTTP(w, req)
	
	// Проверки
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetService(t *testing.T) {
	router, mockDB, _ := setupTestServer()
	
	// Настройка mock
	mockDB.On("QueryRow", mock.AnythingOfType("string"), 1).Return(&database.DB{})
	
	// Создание запроса
	req, _ := http.NewRequest("GET", "/api/v1/services/1", nil)
	w := httptest.NewRecorder()
	
	// Выполнение запроса
	router.ServeHTTP(w, req)
	
	// Проверки
	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetServiceInvalidID(t *testing.T) {
	router, _, _ := setupTestServer()
	
	// Создание запроса с неверным ID
	req, _ := http.NewRequest("GET", "/api/v1/services/invalid", nil)
	w := httptest.NewRecorder()
	
	// Выполнение запроса
	router.ServeHTTP(w, req)
	
	// Проверки
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteService(t *testing.T) {
	router, mockDB, _ := setupTestServer()
	
	// Настройка mock
	mockDB.On("Exec", mock.AnythingOfType("string"), 1).Return(1, nil)
	
	// Создание запроса
	req, _ := http.NewRequest("DELETE", "/api/v1/services/1", nil)
	w := httptest.NewRecorder()
	
	// Выполнение запроса
	router.ServeHTTP(w, req)
	
	// Проверки
	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetAlerts(t *testing.T) {
	router, mockDB, _ := setupTestServer()
	
	// Настройка mock
	mockDB.On("Query", mock.AnythingOfType("string")).Return(&database.DB{}, nil)
	
	// Создание запроса
	req, _ := http.NewRequest("GET", "/api/v1/alerts", nil)
	w := httptest.NewRecorder()
	
	// Выполнение запроса
	router.ServeHTTP(w, req)
	
	// Проверки
	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestResolveAlert(t *testing.T) {
	router, mockDB, _ := setupTestServer()
	
	// Настройка mock
	mockDB.On("Exec", mock.AnythingOfType("string"), 1).Return(1, nil)
	
	// Создание запроса
	req, _ := http.NewRequest("PUT", "/api/v1/alerts/1/resolve", nil)
	w := httptest.NewRecorder()
	
	// Выполнение запроса
	router.ServeHTTP(w, req)
	
	// Проверки
	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetStats(t *testing.T) {
	router, mockDB, _ := setupTestServer()
	
	// Настройка mock
	mockDB.On("QueryRow", mock.AnythingOfType("string")).Return(&database.DB{})
	mockDB.On("QueryRow", mock.AnythingOfType("string")).Return(&database.DB{})
	
	// Создание запроса
	req, _ := http.NewRequest("GET", "/api/v1/stats", nil)
	w := httptest.NewRecorder()
	
	// Выполнение запроса
	router.ServeHTTP(w, req)
	
	// Проверки
	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetServiceChecks(t *testing.T) {
	router, mockDB, _ := setupTestServer()
	
	// Настройка mock
	mockDB.On("Query", mock.AnythingOfType("string"), 1, 100).Return(&database.DB{}, nil)
	
	// Создание запроса
	req, _ := http.NewRequest("GET", "/api/v1/services/1/checks", nil)
	w := httptest.NewRecorder()
	
	// Выполнение запроса
	router.ServeHTTP(w, req)
	
	// Проверки
	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetServiceChecksWithLimit(t *testing.T) {
	router, mockDB, _ := setupTestServer()
	
	// Настройка mock
	mockDB.On("Query", mock.AnythingOfType("string"), 1, 50).Return(&database.DB{}, nil)
	
	// Создание запроса с лимитом
	req, _ := http.NewRequest("GET", "/api/v1/services/1/checks?limit=50", nil)
	w := httptest.NewRecorder()
	
	// Выполнение запроса
	router.ServeHTTP(w, req)
	
	// Проверки
	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}
