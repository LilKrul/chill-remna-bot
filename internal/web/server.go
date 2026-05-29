package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Handlers interface {
	Healthy(ctx context.Context) error

	HandleYooKassaWebhook(ctx context.Context, body []byte) (handled bool, err error)

	HandleCryptoBotWebhook(ctx context.Context, signatureHex string, body []byte) (handled bool, err error)

	HandleRemnawaveWebhook(ctx context.Context, signatureHex string, body []byte) (handled bool, err error)
}

type Server struct {
	log      *slog.Logger
	handlers Handlers
	srv      *http.Server
}

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
