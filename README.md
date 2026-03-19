# BrandRadar

Real-time brand reputation monitoring system. Aggregates mentions from Web, RSS, and Telegram, classifies sentiment via ML, detects anomalies (spikes), and sends alerts.

Built during [Prod-Pobeda 2026](https://prodcontest.ru) hackathon.

## Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Frontend   │────▶│   Backend    │────▶│  ML Service   │
│  React/Vite  │     │   Go / Gin   │     │ FastAPI/BERT  │
│  :8081       │     │  :8080       │     │  :8000        │
└──────────────┘     └──────┬───────┘     └──────────────┘
                            │
                   ┌────────┼────────┐
                   ▼        ▼        ▼
              PostgreSQL   Redis    MinIO
               :5432      :6379    :9000
```

| Component | Stack |
|-----------|-------|
| **Backend** | Go, Gin, Clean Architecture, PostgreSQL (pgx/v5), Redis, Prometheus |
| **Frontend** | React 18, TypeScript, Vite, Tailwind CSS, shadcn/ui, Recharts |
| **ML** | Python, FastAPI, ruBERT (sentiment & relevance), DBSCAN clustering |
| **Monitoring** | Prometheus + Grafana |

## Features

- **Mention collection** from Web (HTML scraping), RSS feeds, and Telegram channels
- **ML classification**: sentiment analysis (positive/neutral/negative, accuracy 0.94) and relevance detection
- **Deduplication** of mentions within project + source
- **Spike alerts** with configurable threshold and cooldown
- **Clustering** of similar mentions via SBERT embeddings + DBSCAN
- **Event log** and **health endpoint** with dependency status
- **Telegram notifications** for anomalies

## Quick Start

### Prerequisites

- Docker & Docker Compose
- (Optional) Go 1.22+, Node.js 20+, Python 3.11+ for local dev

### Run everything

```bash
docker compose up --build
```

This starts all services:

| Service | URL |
|---------|-----|
| Frontend | http://localhost:8081 |
| Backend API | http://localhost:8080/api/v1 |
| Swagger | http://localhost:8080/api/v1/swagger/index.html |
| ML Service | http://localhost:8000/docs |
| Grafana | http://localhost:8083 (admin/admin) |
| Prometheus | http://localhost:8082 |
| MinIO Console | http://localhost:9001 (minioadmin/minioadmin) |

### Run individual components

```bash
# Backend only (with infrastructure)
cd backend && docker compose up --build

# Frontend dev server
cd frontend && npm install && npm run dev

# ML service
cd ml && docker compose up --build
```

## Project Structure

```
BrandRadar/
├── backend/           # Go API server (Clean Architecture)
│   ├── cmd/app/       # Entry point
│   ├── internal/      # Business logic (entity, usecase, repo, controller)
│   ├── migrations/    # PostgreSQL migrations
│   ├── monitoring/    # Prometheus & Grafana configs
│   └── config/        # Application config
├── frontend/          # React SPA
│   └── src/           # Components, pages, API layer
├── ml/                # ML microservice
│   ├── src/           # FastAPI app, models
│   ├── data/          # Training datasets
│   └── eda/           # Exploratory data analysis notebooks
├── docker-compose.yml # Full-stack compose
└── LICENSE
```

## API Overview

```
POST/GET  /api/v1/projects       # Project management
POST/GET  /api/v1/brands         # Brand management
POST/GET  /api/v1/sources        # Data sources (Web, RSS, Telegram)
POST/GET  /api/v1/collector      # Collection control
GET/POST  /api/v1/mentions       # Mentions with ML fields
PUT       /api/v1/mentions/:id/status
POST/GET  /api/v1/alerts         # Spike alert configuration
GET       /api/v1/events         # Event log
GET       /api/v1/health         # Health indicator
```

## Configuration

Copy the example env file and adjust as needed:

```bash
cp backend/.env.example backend/.env
```

Key environment variables:

| Variable | Description |
|----------|-------------|
| `TELEGRAM_BOT_TOKEN` | Telegram bot token for notifications |
| `TELEGRAM_CHAT_ID` | Telegram chat ID for alerts |
| `ML_SENTIMENT_HOST` | ML service URL |
| `PG_PASSWORD` | PostgreSQL password |

## ML Models

| Model | Base | Task | Accuracy |
|-------|------|------|----------|
| Sentiment | ai-forever/ruBert-base | Sentiment classification (pos/neg/neutral) | 0.94 F1: 0.95 |
| Relevance | ai-forever/ruBert-base | Brand relevance detection | ~1.0 |
| Clustering | ai-forever/sbert_large_nlu_ru | Mention grouping (DBSCAN) | — |

Models are stored in MinIO (S3). Training was done using knowledge distillation from Qwen3-235B.

## License

[MIT](LICENSE)
