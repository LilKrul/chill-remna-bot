package app

import (
	"context"
	"time"

	"github.com/go-telegram/bot/models"

	"remnabot/internal/assets"
	"remnabot/internal/i18n"
)

func (a *App) showBroadcast(ctx context.Context, chatID int64) {
	lang := a.lang(chatID)
	count := 0
	if a.store != nil {
		if ids, err := a.store.AllUserIDs(ctx); err == nil {
			count = len(ids)
		}
	}
	a.sendKBSection(ctx, chatID, assets.SectionAdminStats, i18n.T(lang, "bcast.title", count), [][]models.InlineKeyboardButton{
		{btn(i18n.T(lang, "bcast.btn_new"), "bc:new")},
		navBack(lang, "menu:marketing"),
	})
}

func (a *App) onBroadcast(ctx context.Context, chatID int64, val string) {
	lang := a.lang(chatID)
	switch val {
	case "new":
		a.getUI(chatID).adminInput = "bcast"
		a.askInput(ctx, chatID, i18n.T(lang, "bcast.ask"), "menu:broadcast")
	case "send":
		ui := a.getUI(chatID)
		text := ui.broadcastText
		ui.broadcastText = ""
		if text == "" {
			a.showBroadcast(ctx, chatID)
			return
		}
		a.runBroadcast(chatID, text)
		a.sendKB(ctx, chatID, i18n.T(lang, "bcast.started"), [][]models.InlineKeyboardButton{navBack(lang, "menu:marketing")})
	}
}

func (a *App) previewBroadcast(ctx context.Context, chatID int64, text string) {
	lang := a.lang(chatID)
	a.getUI(chatID).broadcastText = text
	count := 0
	if a.store != nil {
		if ids, err := a.store.AllUserIDs(ctx); err == nil {
			count = len(ids)
		}
	}
	a.sendKB(ctx, chatID, i18n.T(lang, "bcast.preview", count)+"\n\n"+text, [][]models.InlineKeyboardButton{
		{btn(i18n.T(lang, "bcast.btn_send", count), "bc:send")},
		navBack(lang, "menu:broadcast"),
	})
}

func (a *App) runBroadcast(adminChat int64, text string) {
	if a.store == nil {
		return
	}
	lang := a.lang(adminChat)
	go func() {
		ctx := context.Background()
		ids, err := a.store.AllUserIDs(ctx)
		if err != nil {
			a.send(ctx, adminChat, i18n.T(lang, "bcast.failed"))
			return
		}
		var sent, failed int
		for _, id := range ids {
			if a.msg.Send(ctx, id, a.applyPremium(text)) != 0 {
				sent++
			} else {
				failed++
			}
			time.Sleep(50 * time.Millisecond)
		}
		id := a.msg.Send(ctx, adminChat, a.applyPremium(i18n.T(lang, "bcast.done", sent, failed)))
		if id != 0 {
			time.AfterFunc(60*time.Second, func() {
				a.msg.Delete(context.Background(), adminChat, id)
			})
		}
	}()
}
