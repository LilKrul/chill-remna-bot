package app

import (
	"context"
	"strings"

	"github.com/go-telegram/bot/models"

	"remnabot/internal/assets"
	"remnabot/internal/i18n"
)

// cabinetURL builds the public URL of the web cabinet from the webhook/public
// settings + the configured path, or "" if no public base is set.
func (a *App) cabinetURL() string {
	a.mu.Lock()
	base := ""
	path := "/cabinet/"
	if a.botCfg != nil {
		base = a.botCfg.Webhook.PublicBaseURL
		if base == "" && a.botCfg.Webhook.Domain != "" {
			base = "https://" + a.botCfg.Webhook.Domain
		}
		if a.botCfg.Cabinet.Path != "" {
			path = a.botCfg.Cabinet.Path
		}
	}
	a.mu.Unlock()
	base = normalizeBaseURL(base)
	if base == "" {
		return ""
	}
	return base + path
}

func (a *App) showCabinetAdmin(ctx context.Context, chatID int64) {
	lang := a.lang(chatID)
	a.mu.Lock()
	on := a.botCfg != nil && a.botCfg.Cabinet.Enabled
	path := "/cabinet/"
	if a.botCfg != nil && a.botCfg.Cabinet.Path != "" {
		path = a.botCfg.Cabinet.Path
	}
	a.mu.Unlock()

	state := i18n.T(lang, "cabinet.off")
	toggle := i18n.T(lang, "cabinet.btn_on")
	if on {
		state = i18n.T(lang, "cabinet.on")
		toggle = i18n.T(lang, "cabinet.btn_off")
	}
	text := i18n.T(lang, "cabinet.title", state, path)
	if url := a.cabinetURL(); url != "" {
		text += "\n\n" + i18n.T(lang, "cabinet.url", url)
	} else {
		text += "\n\n" + i18n.T(lang, "cabinet.no_url")
	}
	text += "\n\n" + i18n.T(lang, "cabinet.steps")

	rows := [][]models.InlineKeyboardButton{
		{btn(toggle, "menu:cabtoggle")},
		{btn(i18n.T(lang, "cabinet.btn_path"), "menu:cabpath")},
		{btn(i18n.T(lang, "btn.back"), "menu:system"), btn(i18n.T(lang, "btn.home"), "menu:home")},
	}
	a.sendKBSection(ctx, chatID, assets.SectionAdminStats, text, rows)
}

func (a *App) toggleCabinet(ctx context.Context, chatID int64) {
	a.mu.Lock()
	if a.botCfg != nil {
		a.botCfg.NormalizeCabinet()
		a.botCfg.Cabinet.Enabled = !a.botCfg.Cabinet.Enabled
	}
	a.mu.Unlock()
	_ = a.saveBotConfig(ctx)
	a.showCabinetAdmin(ctx, chatID)
}

func (a *App) setCabinetPath(ctx context.Context, chatID int64, text string) {
	a.mu.Lock()
	if a.botCfg != nil {
		a.botCfg.Cabinet.Path = strings.TrimSpace(text)
		a.botCfg.NormalizeCabinet()
	}
	a.mu.Unlock()
	_ = a.saveBotConfig(ctx)
	a.showCabinetAdmin(ctx, chatID)
}
