package web

import (
	"context"
	"io"
	"net/http"
	"time"
)

// readAllLimited читает тело запроса с защитой от слишком больших payload'ов
// (CryptoBot/YooKassa посылают единицы килобайт; 256 КиБ — с запасом).
func readAllLimited(r *http.Request, maxBytes int64) ([]byte, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

// handleHealthz — readiness probe. БД и панель опрашиваются через Handlers.
// Возвращаем 200/503 без подробностей, чтобы не утекало в публичный интернет.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if err := s.handlers.Healthy(ctx); err != nil {
		s.log.Warn("healthz: not ready", "err", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("not ready"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// handleYooKassa — POST /webhook/yookassa. YooKassa не подписывает тело,
// поэтому обработчик доверяет не телу, а перепроверяет платёж запросом к API
// YooKassa (GetPayment) — см. HandleYooKassaWebhook.
func (s *Server) handleYooKassa(w http.ResponseWriter, r *http.Request) {
	body, err := readAllLimited(r, 256*1024)
	if err != nil {
		http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	handled, err := s.handlers.HandleYooKassaWebhook(ctx, body)
	if err != nil {
		s.log.Error("yookassa webhook", "err", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	if !handled {
		// 200 на «пропустили» тоже — иначе провайдер начнёт ретраить.
		// Логируем, чтобы можно было увидеть нестандартные ивенты.
		s.log.Info("yookassa webhook ignored")
	}
	w.WriteHeader(http.StatusOK)
}

// handleCryptoBot — POST /webhook/cryptobot. CryptoBot подписывает тело:
// HMAC-SHA256(body, SHA256(token)) в заголовке Crypto-Pay-API-Signature.
// См. https://help.crypt.bot/crypto-pay-api#webhooks.
func (s *Server) handleCryptoBot(w http.ResponseWriter, r *http.Request) {
	sig := r.Header.Get("Crypto-Pay-API-Signature")
	body, err := readAllLimited(r, 256*1024)
	if err != nil {
		http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	handled, err := s.handlers.HandleCryptoBotWebhook(ctx, sig, body)
	if err != nil {
		s.log.Error("cryptobot webhook", "err", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	if !handled {
		s.log.Info("cryptobot webhook ignored")
	}
	w.WriteHeader(http.StatusOK)
}

// handleRemnawave — POST /webhook/remnawave. Панель шлёт HMAC-SHA256 тела
// (ключ — WEBHOOK_SECRET_HEADER из настроек панели) в заголовке
// X-Remnawave-Signature. Сравниваем в постоянном времени (см. Phase 3).
func (s *Server) handleRemnawave(w http.ResponseWriter, r *http.Request) {
	sig := r.Header.Get("X-Remnawave-Signature")
	body, err := readAllLimited(r, 256*1024)
	if err != nil {
		http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	handled, err := s.handlers.HandleRemnawaveWebhook(ctx, sig, body)
	if err != nil {
		s.log.Error("remnawave webhook", "err", err)
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	if !handled {
		s.log.Info("remnawave webhook ignored")
	}
	w.WriteHeader(http.StatusOK)
}
