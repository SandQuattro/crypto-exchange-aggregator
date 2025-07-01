package application

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"crypto-exchange-agg/pkg/logging"

	"crypto-exchange-agg/config"
	"crypto-exchange-agg/internal/currency"
	"crypto-exchange-agg/internal/providers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

const (
	CTX           = "application"
	ResetInterval = 30
)

const (
	SSEHeartbeatInterval   = 10 * time.Second // heartbeat интервал (частый для быстрого обнаружения отключений)
	APIRequestTimeout      = 15 * time.Second // Timeout для API запросов
	HeartbeatMaxRetries    = 3                // Количество попыток для heartbeat
	HeartbeatRetryInterval = 5 * time.Second  // Интервал между попытками
)

type Application struct {
	logger zerolog.Logger

	latestRates map[string]string
	ratesMutex  sync.RWMutex
	clients     map[chan RateUpdate]bool
	clientMutex sync.RWMutex

	client           *providers.CoinMarketCap
	symbols          []currency.Cryptocurrency
	targetCurrencies []string
	provider         string
	interval         time.Duration
	ctx              context.Context
	timer            *time.Timer
}

type RateUpdate struct {
	Timestamp  time.Time         `json:"timestamp"`
	Rates      map[string]string `json:"rates"`
	NextUpdate time.Time         `json:"next_update"`
}

func NewApplication() *Application {
	return &Application{
		latestRates: make(map[string]string),
		clients:     make(map[chan RateUpdate]bool),
	}
}

func (a *Application) Run(ctx context.Context, cfg *config.Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}

	ctx = logging.GetCtxWithScope(logging.GetCtxWithTraceID(ctx), CTX)
	a.logger = logging.GetCtxLogger(ctx)

	// Настройка провайдера и клиента
	a.setupProvider(ctx, cfg)

	// Настройка и запуск Fiber приложения
	app := a.setupFiberApp()
	a.startServer(app, cfg)

	// Запуск основного цикла с таймером
	return a.runMainLoop(ctx, app)
}

// handleSSE обрабатывает Server-Sent Events для стриминга курсов через Fiber.
func (a *Application) handleSSE(c *fiber.Ctx) error {
	appLogger := logging.GetCtxLogger(context.Background())

	a.setupSSEHeaders(c)
	appLogger.Info().Msg("New SSE client connected")

	clientChan := a.registerClient()
	notify := c.Context().Done()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		a.streamToClient(w, clientChan, notify)
	})

	return nil
}

// fetchRates выполняет batch запросы для получения курсов валют.
func (a *Application) fetchRates(
	ctx context.Context,
	client *providers.CoinMarketCap,
	symbols []currency.Cryptocurrency,
	targetCurrencies []string,
	provider string,
	interval time.Duration,
) error {
	logger := logging.GetCtxLogger(ctx)

	g, gCtx := errgroup.WithContext(ctx)

	newRates := make(map[string]string)
	var ratesMutex sync.Mutex

	// Делаем только 2 batch запроса вместо 14 отдельных
	for _, targetCurrency := range targetCurrencies {
		target := targetCurrency

		g.Go(func() error {
			logger.Info().
				Str("target_currency", target).
				Int("symbols_count", len(symbols)).
				Msg("Fetching rates batch")

			rates, rateErr := client.GetRates(gCtx, symbols, target)
			if rateErr != nil {
				logger.Error().
					Err(rateErr).
					Str("target_currency", target).
					Str("provider", provider).
					Msg("Failed to get batch rates")
				return fmt.Errorf("failed to get batch rates for %s: %w", target, rateErr)
			}

			ratesMutex.Lock()
			for symbol, rate := range rates {
				key := fmt.Sprintf("%s_%s", symbol, target)
				newRates[key] = rate

				logger.Info().
					Str("from", symbol).
					Str("to", target).
					Str("rate", rate).
					Str("provider", provider).
					Msg("Rate retrieved successfully")
			}
			ratesMutex.Unlock()

			logger.Info().
				Str("target_currency", target).
				Int("rates_received", len(rates)).
				Msg("Batch rates completed")

			return nil
		})
	}

	if waitErr := g.Wait(); waitErr != nil {
		return fmt.Errorf("batch rates fetch error: %w", waitErr)
	}

	// Обновляем глобальные rates и уведомляем клиентам
	a.ratesMutex.Lock()
	a.latestRates = newRates
	a.ratesMutex.Unlock()

	// Отправляем обновления всем подключенным клиентам
	update := RateUpdate{
		Timestamp:  time.Now(),
		Rates:      newRates,
		NextUpdate: time.Now().Add(interval),
	}

	notifiedClients := a.notifyAllClients(update, logger)

	logger.Info().
		Int("clients_notified", notifiedClients).
		Msg("All rate batches completed successfully")

	return nil
}

// handleFirstClient обрабатывает подключение первого клиента.
func (a *Application) handleFirstClient() {
	a.ratesMutex.RLock()
	hasData := len(a.latestRates) > 0
	a.ratesMutex.RUnlock()

	if hasData {
		return
	}

	a.logger.Info().Msg("First client connected, fetching initial rates...")
	go func() {
		err := a.fetchRates(a.ctx, a.client, a.symbols, a.targetCurrencies, a.provider, a.interval)
		a.resetTimerAfterFetch(err)
	}()
}

// resetTimerAfterFetch сбрасывает таймер после получения данных.
func (a *Application) resetTimerAfterFetch(err error) {
	if a.timer == nil {
		return
	}

	if err != nil {
		a.logger.Error().Err(err).Msg("Initial rate fetch for first client failed")
		a.timer.Reset(ResetInterval * time.Second)
		return
	}

	a.logger.Info().Msg("Initial rates fetched successfully, resetting timer")
	a.timer.Reset(a.interval)
}

// setupProvider настраивает провайдер и клиента для работы с API.
func (a *Application) setupProvider(ctx context.Context, cfg *config.Config) {
	httpClient := &http.Client{
		Timeout: APIRequestTimeout,
	}

	client := &providers.CoinMarketCap{
		Logger: a.logger,
		Client: httpClient,
		Config: cfg,
	}

	a.logger.Info().
		Str("provider", cfg.App.Provider).
		Msg("Using crypto provider")

	symbols := []currency.Cryptocurrency{
		currency.USDT, currency.USDC,
		currency.BTC, currency.ETH, currency.LTC, currency.DOGE,
	}
	targetCurrencies := []string{"USD", "EUR"}

	const intervalSeconds = 1
	interval := time.Minute + intervalSeconds*time.Second

	a.client = client
	a.symbols = symbols
	a.targetCurrencies = targetCurrencies
	a.provider = cfg.Provider
	a.interval = interval
	a.ctx = ctx
}

// setupFiberApp создает и настраивает Fiber приложение.
func (a *Application) setupFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			a.logger.Error().Err(err).Msg("Fiber error")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
		DisableStartupMessage: false,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Cache-Control",
	}))

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))

	// Раздача статических файлов frontend
	app.Static("/", "./static")

	// API Routes
	app.Get("/api/rates/stream", a.handleSSE)

	return app
}

// startServer запускает HTTP сервер в отдельной горутине.
func (a *Application) startServer(app *fiber.App, cfg *config.Config) {
	go func() {
		a.logger.Info().
			Int("port", cfg.HTTP.Port).
			Msg("Starting Fiber HTTP server")

		addr := fmt.Sprintf(":%d", cfg.Port)
		if err := app.Listen(addr); err != nil {
			a.logger.Error().Err(err).Msg("Fiber server error")
		}
	}()
}

// runMainLoop запускает основной цикл приложения с таймером.
func (a *Application) runMainLoop(ctx context.Context, app *fiber.App) error {
	timer := time.NewTimer(a.interval)
	defer timer.Stop()

	a.timer = timer

	a.logger.Info().
		Str("interval", a.interval.String()).
		Msg("Starting periodic rate fetching with timer")

	for {
		select {
		case <-ctx.Done():
			a.logger.Info().Msg("Application context cancelled, stopping...")
			if err := app.Shutdown(); err != nil {
				a.logger.Error().Err(err).Msg("Fiber shutdown error")
			}
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		case <-timer.C:
			a.handleTimerTick(timer)
		}
	}
}

// handleTimerTick обрабатывает срабатывание таймера.
func (a *Application) handleTimerTick(timer *time.Timer) {
	// Проверяем наличие подключенных клиентов перед API запросом
	a.clientMutex.RLock()
	clientsCount := len(a.clients)
	a.clientMutex.RUnlock()

	if clientsCount == 0 {
		a.logger.Info().Msg("No clients connected, skipping API call to save tokens")
		timer.Reset(a.interval)
		return
	}

	a.logger.Info().
		Int("connected_clients", clientsCount).
		Msg("Timer triggered, fetching rates...")

	if err := a.fetchRates(a.ctx, a.client, a.symbols, a.targetCurrencies, a.provider, a.interval); err != nil {
		a.logger.Error().Err(err).Msg("Periodic rate fetch failed")
		timer.Reset(ResetInterval * time.Second)
	} else {
		timer.Reset(a.interval)
	}
}

// setupSSEHeaders настраивает заголовки для Server-Sent Events.
func (a *Application) setupSSEHeaders(c *fiber.Ctx) {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("X-Accel-Buffering", "no") // Отключаем буферизацию для nginx
}

// registerClient регистрирует нового SSE клиента.
func (a *Application) registerClient() chan RateUpdate {
	clientChan := make(chan RateUpdate, 1)

	a.clientMutex.Lock()
	wasEmpty := len(a.clients) == 0
	a.clients[clientChan] = true
	clientsCount := len(a.clients)
	a.clientMutex.Unlock()

	a.logger.Info().
		Int("total_clients", clientsCount).
		Bool("first_client", wasEmpty).
		Msg("SSE client registered")

	if wasEmpty {
		a.handleFirstClient()
	}

	return clientChan
}

// streamToClient обрабатывает стрим данных к клиенту.
func (a *Application) streamToClient(
	w *bufio.Writer,
	clientChan chan RateUpdate,
	notify <-chan struct{},
) {
	defer a.cleanupClient(clientChan)

	if !a.sendInitialRates(w) {
		return
	}

	a.handleClientStream(w, clientChan, notify)
}

// cleanupClient очищает ресурсы клиента.
func (a *Application) cleanupClient(clientChan chan RateUpdate) {
	a.clientMutex.Lock()
	delete(a.clients, clientChan)
	remainingClients := len(a.clients)
	a.clientMutex.Unlock()
	close(clientChan)
	a.logger.Info().
		Int("remaining_clients", remainingClients).
		Msg("SSE client disconnected and cleaned up")
}

// sendInitialRates отправляет начальные данные курсов клиенту.
func (a *Application) sendInitialRates(w *bufio.Writer) bool {
	a.ratesMutex.RLock()
	defer a.ratesMutex.RUnlock()

	if len(a.latestRates) == 0 {
		return true
	}

	update := RateUpdate{
		Timestamp:  time.Now(),
		Rates:      a.latestRates,
		NextUpdate: time.Now().Add(time.Minute + 1*time.Second),
	}

	data, _ := json.Marshal(update)
	if _, err := fmt.Fprintf(w, "event: rates\ndata: %s\n\n", data); err != nil {
		a.logger.Info().Msg("Failed to send initial rates data")
		return false
	}
	if err := w.Flush(); err != nil {
		a.logger.Info().Msg("Failed to flush initial rates data")
		return false
	}

	return true
}

// handleClientStream обрабатывает основной стрим клиента.
func (a *Application) handleClientStream(
	w *bufio.Writer,
	clientChan chan RateUpdate,
	notify <-chan struct{},
) {
	keepAliveTicker := time.NewTicker(SSEHeartbeatInterval)
	defer keepAliveTicker.Stop()

	a.logger.Info().Msg("Starting SSE stream for client")

	for {
		select {
		case <-notify:
			a.logger.Info().Msg("Client disconnected via context.Done()")
			return

		case <-keepAliveTicker.C:
			a.logger.Debug().Msg("Sending heartbeat to client")
			if !a.sendHeartbeat(w) {
				a.logger.Info().Msg("Heartbeat failed, client disconnected")
				return
			}

		case update, ok := <-clientChan:
			if !ok {
				a.logger.Info().Msg("Client channel closed")
				return
			}
			a.logger.Debug().Msg("Sending rate update to client")
			if !a.sendRateUpdate(w, update) {
				a.logger.Info().Msg("Rate update failed, client disconnected")
				return
			}
		}
	}
}

// sendHeartbeat отправляет heartbeat клиенту.
func (a *Application) sendHeartbeat(w *bufio.Writer) bool {
	// Быстрая проверка соединения без retry
	if _, err := w.WriteString("event: heartbeat\ndata: ping\n\n"); err != nil {
		a.logger.Info().Err(err).Msg("Client disconnected (heartbeat write failed)")
		return false
	}

	if err := w.Flush(); err != nil {
		a.logger.Info().Err(err).Msg("Client disconnected (heartbeat flush failed)")
		return false
	}

	return true
}

// sendRateUpdate отправляет обновление курсов клиенту.
func (a *Application) sendRateUpdate(w *bufio.Writer, update RateUpdate) bool {
	data, err := json.Marshal(update)
	if err != nil {
		a.logger.Error().Err(err).Msg("Failed to marshal rate update")
		return true // Продолжаем стрим
	}

	if _, writeErr := fmt.Fprintf(w, "event: rates\ndata: %s\n\n", data); writeErr != nil {
		a.logger.Info().Msg("Client disconnected (write failed)")
		return false
	}

	if flushErr := w.Flush(); flushErr != nil {
		a.logger.Info().Msg("Client disconnected (flush failed)")
		return false
	}

	a.logger.Info().Msg("Sent rate update to SSE client")
	return true
}

// notifyAllClients отправляет обновление всем клиентам (простая версия).
func (a *Application) notifyAllClients(update RateUpdate, _ zerolog.Logger) int {
	a.clientMutex.RLock()
	defer a.clientMutex.RUnlock()

	notifiedCount := 0
	for clientChan := range a.clients {
		select {
		case clientChan <- update:
			notifiedCount++
		default:
			// Канал заблокирован - клиент не готов, пропускаем
			// Отключение будет обнаружено через heartbeat
		}
	}

	return notifiedCount
}

// checkAndCleanupAllClients проверяет состояние всех клиентов и очищает неактивных.
func (a *Application) checkAndCleanupAllClients() { //nolint:unused // Метод для отладки
	a.clientMutex.Lock()
	defer a.clientMutex.Unlock()

	if len(a.clients) == 0 {
		return
	}

	a.logger.Info().
		Int("total_clients", len(a.clients)).
		Msg("Checking client connections")

	disconnectedClients := make([]chan RateUpdate, 0)

	// Проверяем каждый клиент - пытаемся отправить тестовое сообщение
	for clientChan := range a.clients {
		select {
		case <-clientChan:
			// Канал закрыт
			disconnectedClients = append(disconnectedClients, clientChan)
		default:
			// Канал открыт, клиент активен
		}
	}

	// Очищаем отключенных клиентов
	for _, clientChan := range disconnectedClients {
		delete(a.clients, clientChan)
		a.logger.Info().Msg("Cleaned up inactive client channel")
	}

	if len(disconnectedClients) > 0 {
		a.logger.Info().
			Int("cleaned_clients", len(disconnectedClients)).
			Int("active_clients", len(a.clients)).
			Msg("Client cleanup completed")
	}
}
