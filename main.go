package main

import (
	"service-monitor/internal/api"
	"service-monitor/internal/config"
	"service-monitor/internal/database"
	"service-monitor/internal/monitor"
	"service-monitor/internal/logger"
)

func main() {
	// Инициализация логгера
	logger := logger.New()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Ошибка загрузки конфигурации:", err)
	}

	// Подключение к базе данных
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Ошибка подключения к базе данных:", err)
	}
	defer db.Close()

	// Инициализация миграций
	if err := database.Migrate(db); err != nil {
		logger.Fatal("Ошибка миграции базы данных:", err)
	}

	// Создание мониторинга
	monitorService := monitor.NewService(db, logger)

	// Создание API сервера
	server := api.NewServer(cfg, db, monitorService, logger)

	// Запуск мониторинга в фоне
	go monitorService.Start()

	logger.Info("Сервер мониторинга запущен на порту:", cfg.Port)
	
	// Запуск сервера
	if err := server.Run(); err != nil {
		logger.Fatal("Ошибка запуска сервера:", err)
	}
}
