package app

import (
	"context"
	"strings"

	"github.com/go-telegram/bot/models"

	"remnabot/internal/i18n"
	"remnabot/internal/model"
)

func (a *App) showContacts(ctx context.Context, chatID int64) {
	lang := a.lang(chatID)
	a.mu.Lock()
	var c struct{ G, S, T string }
	if a.botCfg != nil {
		c.G = a.botCfg.Contact.GroupURL
		c.S = a.botCfg.Contact.SupportURL
		c.T = a.botCfg.Contact.TermsText
	}
	a.mu.Unlock()
	display := func(v string) string {
		if v == "" {
			return i18n.T(lang, "admin.none")
		}
		return v
	}
	termsStatus := i18n.T(lang, "contacts.terms_off")
	if c.T != "" {
		termsStatus = i18n.T(lang, "contacts.terms_on")
	}
	body := i18n.T(lang, "contacts.title", display(c.G), display(c.S), termsStatus)

	rows := [][]models.InlineKeyboardButton{
		{btn(i18n.T(lang, "contacts.btn_group"), "ctc:group"), btn(i18n.T(lang, "contacts.btn_support"), "ctc:support")},
		{btn(i18n.T(lang, "contacts.btn_terms"), "ctc:terms")},
	}

	if c.G != "" || c.S != "" || c.T != "" {
		rows = append(rows, []models.InlineKeyboardButton{
			btn(i18n.T(lang, "contacts.btn_clear"), "ctc:clear"),
		})
	}
	rows = append(rows, []models.InlineKeyboardButton{
		btn(i18n.T(lang, "btn.back"), "menu:iface"),
		btn(i18n.T(lang, "btn.home"), "menu:home"),
	})
	a.sendKB(ctx, chatID, body, rows)
}

func (a *App) onContacts(ctx context.Context, chatID int64, val string) {
	ui := a.getUI(chatID)
	lang := a.lang(chatID)
	cancel := [][]models.InlineKeyboardButton{{btn(i18n.T(lang, "btn.cancel"), "ctc:cancel")}}
	switch val {
	case "group":
		ui.adminInput = "ctc_group"
		a.sendKB(ctx, chatID, i18n.T(lang, "contacts.ask_group"), cancel)
	case "support":
		ui.adminInput = "ctc_support"
		a.sendKB(ctx, chatID, i18n.T(lang, "contacts.ask_support"), cancel)
	case "terms":
		ui.adminInput = "ctc_terms"
		a.sendKB(ctx, chatID, i18n.T(lang, "contacts.ask_terms"), cancel)
	case "clear":

		a.mu.Lock()
		if a.botCfg != nil {
			a.botCfg.Contact = model.ContactConfig{}
		}
		a.mu.Unlock()
		_ = a.saveBotConfig(ctx)
		a.showContacts(ctx, chatID)
	case "cancel":
		ui.adminInput = ""
		a.showContacts(ctx, chatID)
	}
}

func (a *App) setContact(ctx context.Context, chatID int64, field, raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "-" || raw == "—" {
		raw = ""
	}
	a.mu.Lock()
	if a.botCfg != nil {
		switch field {
		case "group":
			a.botCfg.Contact.GroupURL = raw
		case "support":
			a.botCfg.Contact.SupportURL = raw
		case "terms":
			a.botCfg.Contact.TermsText = raw
		}
	}
	a.mu.Unlock()
	_ = a.saveBotConfig(ctx)
	a.getUI(chatID).adminInput = ""
	a.showContacts(ctx, chatID)
}
