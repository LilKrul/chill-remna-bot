// Package web — HTTP-сервер для приёма входящих вебхуков от платёжных
// провайдеров (YooKassa, CryptoBot) и панели Remnawave (user.* события).
//
// Сервер запускается параллельно с Telegram long-polling'ом и слушает на
// WebhookConfig.ListenAddr. Для боевой эксплуатации перед ним должен стоять
// reverse-proxy (Caddy/nginx/Traefik) с TLS — мы по сути обслуживаем
// http://0.0.0.0:8080 внутри docker'а, а наружу торчит HTTPS.
//
// Маршруты:
//
//	GET  /healthz            — проверка живости (БД + панель), 200 OK / 503
//	POST /webhook/yookassa   — события YooKassa (payment.succeeded и др.)
//	POST /webhook/cryptobot  — события CryptoBot (invoice_paid)
//	POST /webhook/remnawave  — события панели (user.expired, user.created, …)
//
// Каждый хендлер реализуется в отдельном файле (yookassa.go, cryptobot.go,
// remnawave.go); этот файл — только маршрутизация и lifecycle.
package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// Handlers — то, что web-серверу нужно от приложения. Передаём интерфейс,
// чтобы избежать цикла импорта web ↔ app.
type Handlers interface {
	// Healthy сообщает готовность бота (БД открыта, конфиг загружен).
	// Опционально опрашивает панель: при error возвращаем 503.
	Healthy(ctx context.Context) error

	// HandleYooKassaWebhook принимает уже распарсенный JSON ивента.
	// Возвращает (handled, err): handled=true означает, что мы поняли
	// событие (даже если оно дубликат — это идемпотентный success);
	// handled=false — событие игнорируем (например, неподдерживаемый тип).
	HandleYooKassaWebhook(ctx context.Context, body []byte) (handled bool, err error)

	// HandleCryptoBotWebhook — аналогично для CryptoBot.
	HandleCryptoBotWebhook(ctx context.Context, signatureHex string, body []byte) (handled bool, err error)

	// HandleRemnawaveWebhook — события панели. signatureHex — HMAC-SHA256
	// тела по WEBHOOK_SECRET_HEADER (см. docs.rw → webhooks).
	HandleRemnawaveWebhook(ctx context.Context, signatureHex string, body []byte) (handled bool, err error)
}

// Server — обёртка над http.Server.
type Server struct {
	log      *slog.Logger
	handlers Handlers
	srv      *http.Server
}

// New создаёт сервер, не запуская его. addr — host:port (":8080" если пусто).
func New(addr string, h Handlers, log *slog.Logger) *Server {
	if addr == "" {
		addr = ":8080"
	}
	s := &Server{log: log, handlers: h}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /webhook/yookassa", s.handleYooKassa)
	mux.HandleFunc("POST /webhook/cryptobot", s.handleCryptoBot)
	mux.HandleFunc("POST /webhook/remnawave", s.handleRemnawave)
	s.srv = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return s
}

// Run слушает до отмены ctx; при отмене корректно стопится с 5-сек таймаутом.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("HTTP webhook server starting", "addr", s.srv.Addr)
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.srv.Shutdown(shutCtx)
		return nil
	case err := <-errCh:
		return err
	}
}
