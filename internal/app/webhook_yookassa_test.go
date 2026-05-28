package app

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"remnabot/internal/model"
)

// TestYooKassaWebhook_SkipUnknownEvent — события, кроме payment.succeeded,
// должны возвращать handled=false без побочных эффектов.
func TestYooKassaWebhook_SkipUnknownEvent(t *testing.T) {
	a := &App{log: slog.Default()}
	body, _ := json.Marshal(map[string]any{
		"event":  "payment.canceled",
		"object": map[string]any{"id": "pay_x", "paid": false},
	})
	handled, err := a.HandleYooKassaWebhook(context.Background(), body)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if handled {
		t.Errorf("ожидалось handled=false для canceled")
	}
}

// TestYooKassaWebhook_MissingMetadata — без telegram_id / months в metadata
// мы 200-им (handled=true), но ничего не делаем.
func TestYooKassaWebhook_MissingMetadata(t *testing.T) {
	a := &App{log: slog.Default(), botCfg: &model.BotConfig{Installed: true, Language: "ru"}}
	body, _ := json.Marshal(map[string]any{
		"event": "payment.succeeded",
		"object": map[string]any{
			"id": "pay_no_meta", "paid": true, "status": "succeeded",
			"amount":   map[string]string{"value": "150.00", "currency": "RUB"},
			"metadata": map[string]string{},
		},
	})
	handled, err := a.HandleYooKassaWebhook(context.Background(), body)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !handled {
		t.Errorf("ожидалось handled=true (мы съели ивент, чтобы YooKassa не ретраила)")
	}
}

// TestYooKassaWebhook_BadJSON — невалидный JSON => ошибка, чтобы провайдер
// получил 500 и попробовал ретрай. Зловредный POST'ом не подменит платёж.
func TestYooKassaWebhook_BadJSON(t *testing.T) {
	a := &App{log: slog.Default()}
	handled, err := a.HandleYooKassaWebhook(context.Background(), []byte("{bad"))
	if err == nil {
		t.Fatalf("ожидалась ошибка парсинга")
	}
	if handled {
		t.Errorf("при ошибке парсинга handled должен быть false")
	}
}
