# Многоэтапная сборка для оптимизации размера образа
FROM golang:1.21-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache git ca-certificates tzdata

# Установка рабочей директории
WORKDIR /app

# Копирование всех файлов
COPY . .

# Загрузка зависимостей и создание go.sum
RUN go mod download && go mod tidy

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Финальный образ
FROM alpine:latest

# Установка ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates tzdata

# Создание пользователя для безопасности
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Установка рабочей директории
WORKDIR /app

# Копирование бинарного файла из builder
COPY --from=builder /app/main .

# Копирование статических файлов
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Изменение владельца файлов
RUN chown -R appuser:appgroup /app

# Переключение на непривилегированного пользователя
USER appuser

# Открытие порта
EXPOSE 8080

# Команда запуска
CMD ["./main"]
