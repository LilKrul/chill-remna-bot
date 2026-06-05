package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"remnabot/internal/model"
)

type plNotification struct {
	ID            string `json:"id"`
	TransactionID string `json:"transactionId"`
}

func (a *App) HandlePlategaWebhook(ctx context.Context, body []byte) (bool, error) {
	var n plNotification
	if err := json.Unmarshal(body, &n); err != nil {
		return false, fmt.Errorf("platega webhook: bad json: %w", err)
	}
	id := n.ID
	if id == "" {
		id = n.TransactionID
	}
	if id == "" {
		a.log.Warn("platega webhook: empty id")
		return false, nil
	}
	if a.store != nil {
		if done, _ := a.store.PaymentByExtID(ctx, id); done {
			return true, nil
		}
	}
	client := a.plClient()
	if client == nil {
		a.payLog(ctx, model.PayMethodPlatega, id, 0, "error", "клиент Platega не настроен — вебхук нельзя верифицировать")
		a.log.Error("platega webhook: client not configured")
		return true, nil
	}
	tx, err := client.GetTransaction(ctx, id)
	if err != nil {
		a.payLog(ctx, model.PayMethodPlatega, id, 0, "verify_error", "%v", err)
		return false, fmt.Errorf("platega webhook: verify %s: %w", id, err)
	}
	hintTG, _ := parsePlPayload(tx.Payload)
	a.payLog(ctx, model.PayMethodPlatega, id, hintTG, "webhook", "verified via API: status=%s amount=%.2f %s", tx.Status, tx.Amount, tx.Currency)
	if !strings.EqualFold(tx.Status, "CONFIRMED") {
		a.log.Info("platega webhook: not confirmed", "id", id, "status", tx.Status)
		return true, nil
	}
	a.finalizePlatega(ctx, id, tx)
	a.log.Info("platega webhook: finalized", "id", id)
	return true, nil
}
