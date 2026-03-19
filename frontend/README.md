# 🛡️ BrandRadar — Frontend

Система мониторинга упоминаний бренда в реальном времени.

## Возможности

- 📊 Дашборд с метриками и графиками
- 💬 Отслеживание упоминаний бренда
- 🤖 ML-анализ тональности (позитивная / нейтральная / негативная)
- 🔔 Автоматические алерты при всплесках активности
- 🌐 Управление источниками данных (Web, RSS, Telegram)
- 📈 Аналитика по событиям и трендам

## Технологии

| Категория | Стек |
|---|---|
| Фреймворк | Vite + React 18 |
| Язык | TypeScript |
| Стили | Tailwind CSS + shadcn/ui |
| Графики | Recharts |
| Иконки | Lucide React |
| Данные | TanStack React Query |
| Роутинг | React Router v6 |
| HTTP | Axios |
| Тесты | Vitest + Testing Library + Playwright |
| API-типы | openapi-typescript |

## Быстрый старт

### Требования

- [Bun](https://bun.sh) (рекомендуется) или Node.js ≥ 18

### Установка и запуск

```bash
# Установить зависимости
bun install

# Запустить dev-сервер (порт 8081)
bun run dev
```

Приложение будет доступно по адресу: http://localhost:8081

### Переменные окружения

| Переменная | Описание | По умолчанию |
|---|---|---|
| `VITE_API_URL` | URL бэкенда | `http://localhost:8080` |

Для деплоя создайте файл `.env.production`:
```env
VITE_API_URL=http://178.154.219.184/api/v1
```

## Скрипты

| Команда | Описание |
|---|---|
| `bun run dev` | Запуск dev-сервера |
| `bun run build` | Продакшн-сборка |
| `bun run preview` | Превью продакшн-сборки |
| `bun run lint` | Проверка ESLint |
| `bun run type-check` | Проверка TypeScript-типов |
| `bun run test` | Запуск unit-тестов |
| `bun run test:watch` | Тесты в watch-режиме |
| `bun run test:coverage` | Покрытие тестами |

## Структура проекта

```
src/
├── components/          # UI-компоненты
│   ├── dashboard/       # Виджеты дашборда (MentionsChart, SentimentDonut)
│   ├── mentions/        # Фильтры и карточки упоминаний
│   ├── onboarding/      # Онбординг-флоу для новых пользователей
│   ├── settings/        # Диалоги настроек (BrandDialog)
│   ├── shared/          # Общие компоненты (Skeleton, ErrorBanner, EmptyState)
│   └── ui/              # shadcn/ui базовые компоненты
├── contexts/            # React-контексты (ConnectionContext)
├── hooks/               # Кастомные хуки (use-brands, use-mentions, ...)
├── layouts/             # Лейауты (DashboardLayout)
├── lib/                 # Утилиты, API-клиент, моки
│   └── api/             # API-сервисы и типы
├── pages/               # Страницы (Dashboard, Brands, Mentions, ...)
└── test/                # Настройка тестов
```

## Работа с API-типами (openapi-typescript)

Проект использует **openapi-typescript** для автоматической генерации TypeScript-типов из OpenAPI-спецификации бэкенда. Это гарантирует, что типы фронтенда всегда синхронизированы с контрактом API.

### Как это работает

1. Swagger-схема бэкенда хранится в `src/lib/api/swagger3.json`
2. Из неё генерируется файл типов `src/lib/api/types.ts`
3. API-сервисы импортируют типы оттуда:

```typescript
// src/lib/api/brands.ts
import type { components } from "./types";

export type Brand = components["schemas"]["...BrandResponse"];
```

### Обновление типов при изменении бэкенда

```bash
# 1. Скачать свежую схему
curl http://178.154.219.184/api/v1/swagger/doc.json -o src/lib/api/swagger3.json

# 2. Перегенерировать типы
npx openapi-typescript src/lib/api/swagger3.json -o src/lib/api/types.ts

# 3. Проверить что ничего не сломалось
bun run type-check
```

### Файлы API-сервисов

| Файл | Описание |
|---|---|
| `src/lib/api/types.ts` | Автоматически сгенерированные типы (не редактировать вручную) |
| `src/lib/api/brands.ts` | CRUD операции с брендами |
| `src/lib/api/mentions.ts` | Получение упоминаний |
| `src/lib/api/alerts.ts` | Конфигурация и история алертов |
| `src/lib/api/sources.ts` | Управление источниками данных |
| `src/lib/api/events.ts` | Получение событий |
| `src/lib/api/health.ts` | Проверка здоровья бэкенда |
| `src/lib/api.ts` | Базовый Axios-клиент с interceptors |

## Деплой

### Сборка

```bash
bun run build
```

Готовые статические файлы будут в папке `dist/`.

### Nginx (рекомендуется)

```nginx
server {
    listen 8081;
    server_name _;
    root /path/to/frontend/dist;
    index index.html;

    # SPA — все маршруты -> index.html
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Проксирование API на бэкенд
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```
