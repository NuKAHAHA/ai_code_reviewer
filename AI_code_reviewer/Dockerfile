FROM golang:1.25-alpine AS builder

WORKDIR /app

# Модули
COPY go.mod go.sum ./
RUN go mod download

# Весь код
COPY . .

# Сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -o AI_code_reviewer ./cmd/main.go

FROM alpine:3.19

WORKDIR /app

# Копируем бинарь
COPY --from=builder /app/AI_code_reviewer .

# Порты и env по умолчанию
ENV GITLAB_BASE_URL=https://gitlab.com
ENV PORT=8080

EXPOSE 8080

# .env внутри контейнера НЕ используем, работаем через ENV
CMD ["./AI_code_reviewer"]