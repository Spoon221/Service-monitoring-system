# Переменные
BINARY_NAME=service-monitor
DOCKER_IMAGE=service-monitor
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: help build run test clean docker-build docker-run docker-stop lint format

# Помощь
help: ## Показать справку
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'

# Сборка
build: ## Собрать приложение
	@echo "$(GREEN)Сборка приложения...$(NC)"
	go build ${LDFLAGS} -o ${BINARY_NAME} main.go
	@echo "$(GREEN)Готово! Исполняемый файл: ${BINARY_NAME}$(NC)"

build-linux: ## Собрать для Linux
	@echo "$(GREEN)Сборка для Linux...$(NC)"
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-linux main.go
	@echo "$(GREEN)Готово! Исполняемый файл: ${BINARY_NAME}-linux$(NC)"

build-windows: ## Собрать для Windows
	@echo "$(GREEN)Сборка для Windows...$(NC)"
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}.exe main.go
	@echo "$(GREEN)Готово! Исполняемый файл: ${BINARY_NAME}.exe$(NC)"

# Запуск
run: ## Запустить приложение
	@echo "$(GREEN)Запуск приложения...$(NC)"
	go run main.go

run-docker: ## Запустить через Docker Compose
	@echo "$(GREEN)Запуск через Docker Compose...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Приложение доступно по адресу: http://localhost:8080$(NC)"

# Тестирование
test: ## Запустить тесты
	@echo "$(GREEN)Запуск тестов...$(NC)"
	go test -v ./...

test-coverage: ## Запустить тесты с покрытием
	@echo "$(GREEN)Запуск тестов с покрытием...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Отчет о покрытии сохранен в coverage.html$(NC)"

# Очистка
clean: ## Очистить файлы сборки
	@echo "$(GREEN)Очистка файлов сборки...$(NC)"
	rm -f ${BINARY_NAME} ${BINARY_NAME}-linux ${BINARY_NAME}.exe
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Готово!$(NC)"

# Docker
docker-build: ## Собрать Docker образ
	@echo "$(GREEN)Сборка Docker образа...$(NC)"
	docker build -t ${DOCKER_IMAGE}:${VERSION} .
	docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:latest
	@echo "$(GREEN)Docker образ готов: ${DOCKER_IMAGE}:${VERSION}$(NC)"

docker-run: ## Запустить Docker контейнер
	@echo "$(GREEN)Запуск Docker контейнера...$(NC)"
	docker run -d --name ${BINARY_NAME} -p 8080:8080 ${DOCKER_IMAGE}:latest
	@echo "$(GREEN)Контейнер запущен!$(NC)"

docker-stop: ## Остановить Docker контейнер
	@echo "$(GREEN)Остановка Docker контейнера...$(NC)"
	docker stop ${BINARY_NAME} || true
	docker rm ${BINARY_NAME} || true
	@echo "$(GREEN)Контейнер остановлен!$(NC)"

docker-clean: ## Очистить Docker образы
	@echo "$(GREEN)Очистка Docker образов...$(NC)"
	docker rmi ${DOCKER_IMAGE}:latest || true
	docker rmi ${DOCKER_IMAGE}:${VERSION} || true
	@echo "$(GREEN)Docker образы удалены!$(NC)"

# Линтинг и форматирование
lint: ## Запустить линтер
	@echo "$(GREEN)Проверка кода линтером...$(NC)"
	golangci-lint run
	@echo "$(GREEN)Проверка завершена!$(NC)"

format: ## Форматировать код
	@echo "$(GREEN)Форматирование кода...$(NC)"
	go fmt ./...
	@echo "$(GREEN)Код отформатирован!$(NC)"

# Зависимости
deps: ## Установить зависимости
	@echo "$(GREEN)Установка зависимостей...$(NC)"
	go mod download
	@echo "$(GREEN)Зависимости установлены!$(NC)"

deps-update: ## Обновить зависимости
	@echo "$(GREEN)Обновление зависимостей...$(NC)"
	go get -u ./...
	go mod tidy
	@echo "$(GREEN)Зависимости обновлены!$(NC)"

# Разработка
dev: ## Запуск в режиме разработки
	@echo "$(GREEN)Запуск в режиме разработки...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(YELLOW)Air не установлен. Устанавливаем...$(NC)"; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

# Проверка
check: format lint test ## Проверить код (форматирование + линтинг + тесты)

# Установка инструментов разработки
install-tools: ## Установить инструменты разработки
	@echo "$(GREEN)Установка инструментов разработки...$(NC)"
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)Инструменты установлены!$(NC)"

# Информация
info: ## Показать информацию о проекте
	@echo "$(GREEN)Информация о проекте:$(NC)"
	@echo "  Версия: ${VERSION}"
	@echo "  Время сборки: ${BUILD_TIME}"
	@echo "  Go версия: $(shell go version)"
	@echo "  Архитектура: $(shell go env GOOS)/$(shell go env GOARCH)"

# Миграции базы данных
db-migrate: ## Запустить миграции БД
	@echo "$(GREEN)Запуск миграций базы данных...$(NC)"
	@echo "$(YELLOW)Миграции выполняются автоматически при запуске приложения$(NC)"

# Создание релиза
release: clean build-linux ## Создать релиз
	@echo "$(GREEN)Создание релиза...$(NC)"
	tar -czf ${BINARY_NAME}-${VERSION}-linux-amd64.tar.gz ${BINARY_NAME}-linux
	@echo "$(GREEN)Релиз создан: ${BINARY_NAME}-${VERSION}-linux-amd64.tar.gz$(NC)"

# Мониторинг
logs: ## Показать логи Docker контейнера
	@echo "$(GREEN)Логи контейнера:$(NC)"
	docker logs -f ${BINARY_NAME}

# Остановка всех сервисов
stop-all: docker-stop ## Остановить все сервисы
	@echo "$(GREEN)Остановка всех сервисов...$(NC)"
	docker-compose down
	@echo "$(GREEN)Все сервисы остановлены!$(NC)"
