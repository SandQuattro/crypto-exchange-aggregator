# Crypto Exchange Aggregator

## 📈 Описание

Real-time агрегатор курсов криптовалют с веб-интерфейсом и SSE стримингом. Проект построен на Clean Architecture с использованием Go backend и React frontend.

## ✨ Реализованные функции

- 🔄 **Real-time обновления** через Server-Sent Events (SSE)
- 📊 **6 криптовалют**: USDT, USDC, BTC, ETH, LTC, DOGE
- 💱 **2 валюты**: USD и EUR курсы
- 📈 **Индикаторы изменения цен** с цветовой кодировкой (зеленый ↗ / красный ↘)
- 🚀 **Batch API запросы** (2 вместо 14 отдельных)
- ⚡ **Умная оптимизация токенов** - API вызовы только при подключенных клиентах
- 💓 **Heartbeat система** для отслеживания соединений (10 сек)
- 🐳 **Docker контейнеризация** с multi-stage build
- 📱 **Адаптивный веб-интерфейс** с таймером обновлений

## 🛠 Технический стек

### Backend

- **Go 1.24** с Fiber framework
- **CoinMarketCap Professional API**
- **zerolog** для structured logging  
- **golangci-lint** для качества кода
- **Docker** для контейнеризации

### Frontend

- **React 18** с react-scripts
- **SSE (EventSource)** для real-time данных
- **CSS3** с анимациями и адаптивностью
- **Static serving** из Go backend

### DevOps

- **Docker Compose** для локальной разработки
- **Multi-stage Dockerfile** для оптимизации размера
- **Makefile** для автоматизации команд

## 🏗 Архитектура

Проект следует принципам **Clean Architecture**:

```
├── cmd/aggregator/           # Точка входа приложения
├── internal/
│   ├── application/         # Бизнес-логика и SSE обработка
│   ├── currency/           # Доменные модели криптовалют
│   └── providers/          # Внешние API провайдеры
├── pkg/logging/            # Переиспользуемые утилиты
├── config/                 # Конфигурация приложения
├── frontend/               # React веб-интерфейс
└── static/                 # Собранные статические файлы
```

### Ключевые компоненты

- **SSE Connection Management**: Context.Done() + heartbeat для отслеживания клиентов
- **Batch API Processing**: Конкурентные запросы через errgroup
- **Smart Token Usage**: API вызовы только при активных соединениях
- **Modular Functions**: Все функции <100 строк, низкая cognitive complexity

## 🚀 Быстрый старт

### Предварительные требования

- Docker и Docker Compose
- CoinMarketCap API ключ

### Запуск

1. **Клонируйте репозиторий:**

   ```bash
   git clone <repository-url>
   cd crypto-exchange-aggregator
   ```

2. **Настройте конфигурацию:**

   ```bash
   # Добавьте ваш CoinMarketCap API ключ в config/config.json
   {
     "key": "your-coinmarketcap-api-key",
     "app": {
       "provider": "coinmarketcap"
     },
     "http": {
       "port": 8080
     }
   }
   ```

3. **Запустите приложение:**

   ```bash
   make dc  # docker-compose up --build
   ```

4. **Откройте браузер:**

   ```
   http://localhost:8080
   ```

## 🔧 Команды разработки

```bash
# Линтинг Go кода
make lint

# Запуск Docker Compose
make dc

# Сборка backend
go build -o bin/aggregator cmd/aggregator/main.go

# Сборка frontend
cd frontend && npm run build
```

## 📡 API Документация

### SSE Endpoint

**GET** `/api/rates/stream`

**Response Events:**

- `rates` - Обновления курсов криптовалют
- `heartbeat` - Проверка соединения (каждые 10 сек)

**Пример rate события:**

```json
{
  "timestamp": "2025-06-30T19:00:00Z",
  "rates": {
    "BTC_USD": "107279.94968997",
    "BTC_EUR": "91102.13327672",
    "ETH_USD": "2483.78102170",
    "ETH_EUR": "2109.22684362"
  },
  "next_update": "2025-06-30T19:01:05Z"
}
```

## 📊 Мониторинг

### Логирование

Проект использует structured logging с zerolog:

```bash
# Примеры логов
INF New SSE client connected
INF Timer triggered, fetching rates... connected_clients=1
INF All rate batches completed successfully clients_notified=1
INF Client disconnected via context.Done() remaining_clients=0
```

### Метрики

- **API интервал**: ~65 секунд (1 мин + 1 сек)
- **Heartbeat**: 10 секунд
- **API timeout**: 15 секунд
- **Retry интервал**: 30 секунд при ошибках

## 🎯 Планы развития

- [ ] **WebSocket поддержка** как альтернатива SSE
- [ ] **Больше криптовалют** и фиатных валют  
- [ ] **Исторические данные** и графики
- [ ] **Алерты** на изменения цен
- [ ] **API ключи через переменные окружения**
- [ ] **Prometheus метрики** для мониторинга
- [ ] **Unit тесты** для всех компонентов

## 📝 Статус проекта

✅ **Стабильная версия** - все основные функции реализованы  
✅ **0 issues** в golangci-lint  
✅ **SSE соединения** работают корректно  
✅ **Clean Architecture** соблюдается  
🧪 **Готово к production** после финального тестирования

## 🤝 Разработка

При разработке соблюдайте:

- **Go Code Style**: gofumpt + golangci-lint
- **Commit Convention**: Conventional Commits
- **Architecture**: Clean Architecture принципы
- **Testing**: Покрытие критического кода тестами

---

**🚀 Готово к использованию!** Запустите `make dc` и откройте <http://localhost:8080>
