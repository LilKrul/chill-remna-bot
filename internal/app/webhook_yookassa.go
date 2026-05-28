package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-telegram/bot/models"

	"remnabot/internal/i18n"
	"remnabot/internal/model"
	"remnabot/internal/storage"
)

// ykNotification — формат POST-уведомления YooKassa.
// См. https://yookassa.ru/developers/using-api/webhooks
// Нам интересны только успешные платежи (event = payment.succeeded);
// payment.canceled приходит, когда юзер закрыл форму оплаты — её
// игнорируем, юзер сам нажмёт «Оплатить» заново.
type ykNotification struct {
	Type   string `json:"type"`  // notification
	Event  string `json:"event"` // payment.succeeded | payment.canceled | refund.succeeded ...
	Object struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Paid   bool   `json:"paid"`
		Amount struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"amount"`
		Metadata map[string]string `json:"metadata"`
	} `json:"object"`
}

// HandleYooKassaWebhook — обработчик /webhook/yookassa (Phase 2).
//
// Идемпотентность:
//   - до вызова finalizePurchase проверяем PaymentByExtID(payment.id):
//     если запись уже есть — возвращаем (handled=true, nil) без побочных
//     эффектов (юзеру повторно ничего не шлём, чтобы не спамить).
//   - на гонке (две одновременные доставки) AddPayment вернёт
//     storage.ErrDuplicateExtID благодаря UNIQUE-индексу из миграции 0014.
//
// Polling-флоу через кнопку «Проверить» остаётся как fallback на случай,
// если бот за reverse-proxy недоступен или вебхук «не доехал».
func (a *App) HandleYooKassaWebhook(ctx context.Context, body []byte) (bool, error) {
	var n ykNotification
	if err := json.Unmarshal(body, &n); err != nil {
		return false, fmt.Errorf("yookassa webhook: bad json: %w", err)
	}
	if n.Event != "payment.succeeded" {
		// Возвращаем handled=false для логирования, но без ошибки — YooKassa
		// получит 200 OK и не будет ретраить (что и нужно для canceled/refund).
		a.log.Info("yookassa webhook: skipping event", "event", n.Event, "id", n.Object.ID)
		return false, nil
	}
	if n.Object.ID == "" || !n.Object.Paid {
		a.log.Warn("yookassa webhook: empty id or unpaid", "id", n.Object.ID, "paid", n.Object.Paid)
		return false, nil
	}

	// Идемпотентность: уже зачтено — выходим тихо.
	if a.store != nil {
		if done, _ := a.store.PaymentByExtID(ctx, n.Object.ID); done {
			a.log.Info("yookassa webhook: duplicate (already finalized)", "id", n.Object.ID)
			return true, nil
		}
	}

	// Метаданные кладём при создании платежа в yookassa.Client.CreatePayment.
	chatID, _ := strconv.ParseInt(n.Object.Metadata["telegram_id"], 10, 64)
	months, _ := strconv.Atoi(n.Object.Metadata["months"])
	if chatID == 0 || months == 0 {
		// Без метаданных не понимаем, кому и что зачислять. Возвращаем 200
		// (не ретраить), но фиксируем как ошибку в логах админа.
		a.log.Error("yookassa webhook: missing metadata",
			"id", n.Object.ID, "telegram_id_raw", n.Object.Metadata["telegram_id"], "months_raw", n.Object.Metadata["months"])
		return true, nil
	}

	amount := n.Object.Amount.Value + " " + n.Object.Amount.Currency
	link, err := a.finalizePurchase(ctx, chatID, months, model.PayMethodYooKassa, amount, n.Object.ID)
	if err != nil {
		// ErrDuplicateExtID — гонка с другой доставкой того же платежа.
		// Не считаем ошибкой: тот webhook завершит флоу.
		if errors.Is(err, storage.ErrDuplicateExtID) {
			a.log.Info("yookassa webhook: race lost (other delivery won)", "id", n.Object.ID)
			return true, nil
		}
		return false, fmt.Errorf("finalize yookassa %s: %w", n.Object.ID, err)
	}

	// Пушим юзеру факт оплаты с прямой ссылкой на подписку.
	lang := a.lang(chatID)
	a.notifyKB(ctx, chatID, i18n.T(lang, "yk.paid_ok", link), [][]models.InlineKeyboardButton{
		{btn(i18n.T(lang, "btn.mysubs"), "menu:mysubs")},
	})
	a.log.Info("yookassa webhook: payment finalized", "id", n.Object.ID, "chat_id", chatID, "months", months)
	return true, nil
}
