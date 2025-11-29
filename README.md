AI Code Review System
Общая информация

Этот проект представляет собой систему автоматической проверки кода с использованием двух микросервисов: AI Reviewer и CodeSage. Система интегрируется с GitLab для автоматической проверки Merge Requests, анализа диффов, выявления ошибок в коде и генерации отзывов с помощью AI.

Структура проекта:

AI Reviewer Service

Слушает Webhook-события GitLab (Merge Request Hook).

Получает дифф кода через GitLab API.

Отправляет дифф в CodeSage для статического анализа.

Генерирует отзыв с использованием LLaMA через Ollama API.

Оставляет комментарий в Merge Request.

CodeSage Analysis Service

Анализирует Go-код (DIFF) с использованием:

AST (Abstract Syntax Tree) для синтаксического анализа.

golangci-lint для проверки на линтинговые ошибки.

Возвращает JSON с найденными ошибками.

Архитектура

Проект использует микросервисную архитектуру, состоящую из двух основных сервисов:

AI Reviewer (реализован на Go)

CodeSage (реализован на Go)

Оба сервиса работают в Docker-контейнерах и взаимодействуют через HTTP API.

Поток данных (Pipeline)

GitLab MR Event

AI Reviewer (Go)

Получает дифф с GitLab API

Отправляет дифф в CodeSage для анализа

CodeSage (Go)

Выполняет AST и Lint анализ

AI Reviewer

Формирует запрос для LLaMA

Отправляет запрос в Ollama

LLaMA (через Ollama API)

Получает запрос и генерирует текст отзыва

AI Reviewer

Получает отзыв от LLaMA

Публикует комментарий в Merge Request в GitLab

Управление окружением
Конфигурация для AI Reviewer

Создайте файл .env с следующими переменными:

GITLAB_BASE_URL=https://gitlab.com
GITLAB_TOKEN=YOUR_TOKEN
GITLAB_WEBHOOK_SECRET=YOUR_SECRET
PORT=8080
CODESAGE_URL=http://codesage:8081/analyze
LLAMA_URL=http://host.docker.internal:11434/api/generate


host.docker.internal позволяет контейнеру видеть Ollama, запущенного на Windows.

Docker Compose

Для запуска системы используйте Docker Compose:

services:
  codesage:
    build: ./codesage
    container_name: codesage
    ports:
      - "8081:8081"
    networks:
      - ai-net
    restart: unless-stopped

  ai-reviewer:
    build: ./AI_code_reviewer
    container_name: ai-reviewer
    depends_on:
      - codesage
    environment:
      GITLAB_BASE_URL: "https://gitlab.com"
      GITLAB_TOKEN: "${GITLAB_TOKEN}"
      GITLAB_WEBHOOK_SECRET: "${GITLAB_WEBHOOK_SECRET}"
      PORT: "8080"
      CODESAGE_URL: "http://codesage:8081/analyze"
      LLAMA_URL: "http://host.docker.internal:11434/api/generate"
    ports:
      - "8080:8080"
    networks:
      - ai-net
    restart: unless-stopped

networks:
  ai-net:
    driver: bridge

Как запустить проект
1. Установите Ollama

Скачайте Ollama с официального сайта
.

Проверьте установку:

ollama list
ollama pull llama3.1

2. Запустите сервер Ollama

Запустите Ollama сервер:

ollama serve


Тестирование:

curl http://localhost:11434/api/generate -d '{"model":"llama3.1","prompt":"hello"}'

3. Запустите сервисы через Docker

В корне проекта выполните:

docker compose up --build -d

4. Настройте Webhook в GitLab

Перейдите в Settings → Webhooks:

URL: http://<ngrok-url>/gitlab/webhook

Secret Token: Значение из .env

Trigger: Merge Request Hook

5. Создайте Merge Request

Создайте изменения и создайте Merge Request.

Система скачает дифф, передаст его в CodeSage для анализа, LLaMA сгенерирует отзыв и добавит комментарий в Merge Request.

Проверка работоспособности
Проверка CodeSage

Тестирование сервиса CodeSage:

curl -X POST http://localhost:8081/analyze \
  -H "Content-Type: application/json" \
  -d '{"diff": "+++ b/main.go\n+func main() { bad code }"}'


Ожидаются ошибки AST.

Проверка AI Reviewer

Проверьте логи контейнера для AI Reviewer:

docker logs -f ai-reviewer


Вы должны увидеть логи, которые подтверждают успешную работу:

codesage lint_issues=...

codesage ast_issues=...

LLAMA review generated

comment posted

Текущие возможности проекта

Ловит события Merge Request в GitLab.

Правильно обрабатывает DIFF.

Отправляет дифф в CodeSage для анализа.

Выполняет AST-анализ для поиска синтаксических ошибок.

Проверяет код с помощью golangci-lint.

Генерирует и публикует комментарии с помощью LLaMA.

Стабильно работает в Docker.

Интегрирован с Ollama для генерации текстов отзывов.
