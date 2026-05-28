package app

import (
	"context"
	"strings"

	"github.com/go-telegram/bot/models"

	"remnabot/internal/i18n"
)

// --- админ: «Замена саб-домена для подписок» ---
//
// Поток:
//   1) showSubdomain — карточка с текущим состоянием (домен или «выключено»),
//      кнопки «✏️ Изменить» / «❌ Сбросить» / «Назад» / «Главная».
//   2) askSubdomain — переводит uiState в ожидание текста, текстовый ввод
//      админа (handleAdminText) подхватит и вызовет setSubdomain(text).
//   3) Если админ присылает «-» или «—», override очищается.

func (a *App) showSubdomain(ctx context.Context, chatID int64) {
	lang := a.lang(chatID)
	cur := a.subOverride()
	statusKey := "subdomain.off"
	if cur != "" {
		statusKey = "subdomain.on"
	}
	rows := [][]models.InlineKeyboardButton{
		{btn(i18n.T(lang, "subdomain.btn_change"), "subd:edit")},
	}
	if cur != "" {
		rows = append(rows, []models.InlineKeyboardButton{
			btn(i18n.T(lang, "subdomain.btn_clear"), "subd:clear"),
		})
	}
	rows = append(rows, []models.InlineKeyboardButton{
		btn(i18n.T(lang, "btn.back"), "menu:manage"),
		btn(i18n.T(lang, "btn.home"), "menu:home"),
	})

	display := cur
	if display == "" {
		display = i18n.T(lang, "admin.none")
	}
	a.sendKB(ctx, chatID, i18n.T(lang, "subdomain.title",
		i18n.T(lang, statusKey), display), rows)
}

// askSubdomain переводит ожидание ввода: handleAdminText распознаёт adminInput=="subdomain".
func (a *App) askSubdomain(ctx context.Context, chatID int64) {
	lang := a.lang(chatID)
	a.getUI(chatID).adminInput = "subdomain"
	a.getUI(chatID).priceMonths = 0 // не пересекается, но обнулим для чистоты
	a.sendKB(ctx, chatID, i18n.T(lang, "subdomain.ask"), [][]models.InlineKeyboardButton{
		{btn(i18n.T(lang, "btn.cancel"), "subd:cancel")},
	})
}

// setSubdomain сохраняет новое значение (или сбрасывает при пустом/«-»).
func (a *App) setSubdomain(ctx context.Context, chatID int64, raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "-" || raw == "—" {
		raw = ""
	}
	// Принимаем «https://x.io» или просто «x.io» — нормализуем до host.
	host := extractHost(raw)
	if host == "" && raw != "" {
		host = raw // fallback: оставим как ввёл (если уже host:port)
	}
	a.mu.Lock()
	if a.botCfg != nil {
		a.botCfg.SubscriptionDomain = host
	}
	a.mu.Unlock()
	_ = a.saveBotConfig(ctx)
	a.getUI(chatID).adminInput = ""
	a.showSubdomain(ctx, chatID)
}

// onSubdomain — диспетчер callback'ов "subd:*".
func (a *App) onSubdomain(ctx context.Context, chatID int64, val string) {
	switch val {
	case "edit":
		a.askSubdomain(ctx, chatID)
	case "clear":
		a.setSubdomain(ctx, chatID, "")
	case "cancel":
		a.getUI(chatID).adminInput = ""
		a.showSubdomain(ctx, chatID)
	}
}
