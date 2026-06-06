package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"remnabot/internal/i18n"
	"remnabot/internal/model"
)

func (a *App) payLog(ctx context.Context, method, extID string, telegramID int64, stage, format string, args ...any) {
	detail := fmt.Sprintf(format, args...)
	a.log.Info("paylog", "method", method, "ext_id", extID, "tg_id", telegramID, "stage", stage, "detail", detail)
	if a.store == nil {
		return
	}
	_ = a.store.AddPayLog(ctx, &model.PayLogEntry{
		ExtID:      extID,
		TelegramID: telegramID,
		Method:     method,
		Stage:      stage,
		Detail:     detail,
	})
}

func (a *App) adminSendPayLog(ctx context.Context, chatID int64, query string) {
	lang := a.lang(chatID)
	query = strings.TrimSpace(query)
	if query == "" || a.store == nil {
		a.sendHome(ctx, chatID, i18n.T(lang, "paylog.empty", query))
		return
	}
	var tg int64
	if n, err := strconv.ParseInt(query, 10, 64); err == nil {
		tg = n
	}
	entries, err := a.store.PayLogs(ctx, query, tg, 2000)
	if err != nil {
		a.sendHome(ctx, chatID, "❌ "+err.Error())
		return
	}
	if len(entries) == 0 {
		a.sendHome(ctx, chatID, i18n.T(lang, "paylog.empty", query))
		return
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "payment log · query=%s · entries=%d · generated=%s\n\n", query, len(entries), time.Now().UTC().Format(time.RFC3339))
	for _, e := range entries {
		fmt.Fprintf(&sb, "%s [%s] ext=%s tg=%d %s: %s\n", e.CreatedAt, e.Method, e.ExtID, e.TelegramID, e.Stage, e.Detail)
	}
	name := "paylog_" + sanitizeFileName(query) + ".log"
	a.msg.SendDocument(ctx, chatID, name, []byte(sb.String()), i18n.T(lang, "paylog.caption", query, len(entries)))
}

func sanitizeFileName(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "query"
	}
	return b.String()
}
