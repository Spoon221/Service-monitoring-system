# 📊 Система мониторинга сервисов

Современная система мониторинга сервисов, написанная на Go с красивым веб-интерфейсом. Позволяет отслеживать доступность веб-сервисов, получать уведомления о проблемах и анализировать статистику uptime.

## ✨ Возможности

- 🔍 **Мониторинг сервисов** - автоматическая проверка доступности по HTTP/HTTPS
- 📈 **Real-time дашборд** - обновление данных в реальном времени через WebSocket
- 🚨 **Система алертов** - автоматические уведомления при недоступности сервисов
- 📊 **Статистика и графики** - детальная аналитика uptime и времени отклика
- 🎨 **Современный UI** - адаптивный дизайн с поддержкой мобильных устройств
- 🔧 **REST API** - полный набор endpoints для интеграции
- 🐳 **Docker поддержка** - готовые контейнеры для быстрого развертывания

## 🏗️ Архитектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API   │    │   PostgreSQL    │
│   (HTML/CSS/JS) │◄──►│   (Go/Gin)      │◄──►│   Database      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │
         │              ┌─────────────────┐
         └──────────────►│   WebSocket     │
                        │   (Real-time)   │
                        └─────────────────┘
```

## 🚀 Быстрый старт

### Вариант 1: Docker Compose (рекомендуется)

1. **Клонируйте репозиторий:**
```bash
git clone https://github.com/Spoon221/service-monitor.git
cd service-monitor
```

2. **Запустите приложение:**
```bash
docker-compose up -d
```

3. **Откройте браузер:**
```
http://localhost:8080
```

### Вариант 2: Локальная установка

1. **Установите зависимости:**
   - Go 1.21+
   - PostgreSQL 12+

2. **Настройте базу данных:**
```sql
CREATE DATABASE service_monitor;
```

3. **Скопируйте конфигурацию:**
```bash
cp env.example .env
# Отредактируйте .env файл под ваши настройки
```

4. **Запустите приложение:**
```bash
go mod download
go run main.go
```

## 📖 API Документация

### Сервисы

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/v1/services` | Получить список всех сервисов |
| POST | `/api/v1/services` | Создать новый сервис |
| GET | `/api/v1/services/:id` | Получить информацию о сервисе |
| PUT | `/api/v1/services/:id` | Обновить сервис |
| DELETE | `/api/v1/services/:id` | Удалить сервис |

### Алерты

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/v1/alerts` | Получить список алертов |
| PUT | `/api/v1/alerts/:id/resolve` | Разрешить алерт |

### Статистика

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/v1/stats` | Получить общую статистику |

### WebSocket

| Endpoint | Описание |
|----------|----------|
| `/api/v1/ws` | Real-time обновления |

## 🎯 Примеры использования

### Добавление сервиса через API

```bash
curl -X POST http://localhost:8080/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Google",
    "url": "https://www.google.com",
    "check_interval": 30,
    "timeout": 10
  }'
```

### Получение статистики

```bash
curl http://localhost:8080/api/v1/stats
```

## 🖼️ Скриншоты

### Главный дашборд
![Главный дашборд](./images/dashboard.png)

### Добавление сервиса
![Добавление сервиса](./images/add-service.png)

### Детали сервиса
![Детали сервиса](./images/service-details.png)

## ⚙️ Конфигурация

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `DATABASE_URL` | URL подключения к PostgreSQL | `postgres://postgres:password@localhost:5432/service_monitor?sslmode=disable` |
| `PORT` | Порт веб-сервера | `8080` |
| `LOG_LEVEL` | Уровень логирования | `info` |
| `CHECK_INTERVAL` | Интервал проверки сервисов (сек) | `30` |

## 🧪 Тестирование

### Запуск тестов
```bash
go test ./...
```

### Тестирование API
```bash
# Тест создания сервиса
curl -X POST http://localhost:8080/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Service", "url": "https://httpbin.org/status/200"}'

# Тест получения сервисов
curl http://localhost:8080/api/v1/services
```

## 📁 Структура проекта

```
service-monitor/
├── main.go                 # Точка входа приложения
├── go.mod                  # Зависимости Go
├── Dockerfile              # Конфигурация Docker
├── docker-compose.yml      # Docker Compose конфигурация
├── env.example             # Пример конфигурации
├── README.md               # Документация
├── internal/               # Внутренние пакеты
│   ├── api/               # REST API и WebSocket
│   ├── config/            # Конфигурация
│   ├── database/          # Работа с БД
│   ├── logger/            # Логирование
│   ├── models/            # Модели данных
│   └── monitor/           # Логика мониторинга
├── static/                # Статические файлы
│   ├── css/              # Стили
│   └── js/               # JavaScript
└── templates/             # HTML шаблоны
```

## 🔧 Разработка

### Требования
- Go 1.21+
- PostgreSQL 12+
- Node.js (для сборки фронтенда, опционально)

### Локальная разработка
```bash
# Установка зависимостей
go mod download

# Запуск в режиме разработки
go run main.go

# Запуск с hot reload (требует air)
air
```

### Сборка
```bash
# Сборка для Linux
GOOS=linux GOARCH=amd64 go build -o service-monitor main.go

# Сборка Docker образа
docker build -t service-monitor .
```

## 🤝 Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для новой функции (`git checkout -b feature/amazing-feature`)
3. Зафиксируйте изменения (`git commit -m 'Add amazing feature'`)
4. Отправьте в ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## 🚀 Roadmap

- [ ] Email уведомления
- [ ] Telegram бот интеграция
- [ ] Метрики Prometheus
- [ ] Grafana дашборды
- [ ] Множественные проверки (TCP, DNS, etc.)
- [ ] Географическое распределение проверок
- [ ] API ключи для аутентификации
- [ ] Роли и права доступа
- [ ] Экспорт данных
- [ ] Интеграция с внешними системами

---