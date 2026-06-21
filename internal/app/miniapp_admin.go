package app

import (
	"context"
	"strings"

	"github.com/go-telegram/bot/models"

	"remnabot/internal/assets"
	"remnabot/internal/i18n"
)

// miniAppURL builds the public URL of the Mini App from the webhook/public
// settings, or "" if no public base is configured yet.
func (a *App) miniAppURL() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.botCfg == nil {
		return ""
	}
	base := strings.TrimRight(a.botCfg.Webhook.PublicBaseURL, "/")
	if base == "" && a.botCfg.Webhook.Domain != "" {
		base = "https://" + a.botCfg.Webhook.Domain
	}
	if base == "" {
		return ""
	}
	return base + "/miniapp/"
}

func (a *App) showMiniAppAdmin(ctx context.Context, chatID int64) {
	lang := a.lang(chatID)
	a.mu.Lock()
	on := a.botCfg != nil && a.botCfg.MiniApp.Enabled
	a.mu.Unlock()

	state := i18n.T(lang, "miniapp.off")
	toggle := i18n.T(lang, "miniapp.btn_on")
	if on {
		state = i18n.T(lang, "miniapp.on")
		toggle = i18n.T(lang, "miniapp.btn_off")
	}
	url := a.miniAppURL()
	text := i18n.T(lang, "miniapp.title", state)
	if url != "" {
		text += "\n\n" + i18n.T(lang, "miniapp.url", url)
	} else {
		text += "\n\n" + i18n.T(lang, "miniapp.no_url")
	}
	text += "\n\n" + i18n.T(lang, "miniapp.steps")

	rows := [][]models.InlineKeyboardButton{
		{btn(toggle, "menu:miniapptoggle")},
		{btn(i18n.T(lang, "btn.back"), "menu:system"), btn(i18n.T(lang, "btn.home"), "menu:home")},
	}
	a.sendKBSection(ctx, chatID, assets.SectionAdminStats, text, rows)
}

func (a *App) toggleMiniApp(ctx context.Context, chatID int64) {
	a.mu.Lock()
	if a.botCfg != nil {
		a.botCfg.MiniApp.Enabled = !a.botCfg.MiniApp.Enabled
		a.botCfg.MiniApp.Init = true
	}
	a.mu.Unlock()
	_ = a.saveBotConfig(ctx)
	a.showMiniAppAdmin(ctx, chatID)
}

// miniAppButtonRow returns a web_app launch button row for the Mini App, or nil
// when the feature is disabled or no public URL is configured.
func (a *App) miniAppButtonRow(lang string) []models.InlineKeyboardButton {
	a.mu.Lock()
	on := a.botCfg != nil && a.botCfg.MiniApp.Enabled
	a.mu.Unlock()
	if !on {
		return nil
	}
	url := a.miniAppURL()
	if url == "" {
		return nil
	}
	return []models.InlineKeyboardButton{{Text: i18n.T(lang, "btn.open_app"), WebApp: &models.WebAppInfo{URL: url}}}
}
