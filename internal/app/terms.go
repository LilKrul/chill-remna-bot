package app

import (
	"context"
	"time"

	"github.com/go-telegram/bot/models"

	"remnabot/internal/i18n"
)

// --- пользовательское соглашение перед первой покупкой ---
//
// Логика:
//   • Если в Contact.TermsText есть текст и users.terms_accepted_at у этого
//     юзера пустое — перед выбором тарифа бот показывает текст соглашения и
//     просит нажать «✅ Принимаю». До этого момента флоу покупки не идёт.
//   • После принятия в users.terms_accepted_at пишется ISO-время и юзер
//     попадает на showPlans без повторного запроса.
//   • Если TermsText пуст — соглашение полностью отключено.

// termsRequired возвращает (textIfShow, true) если перед покупкой надо
// показать соглашение; (",", false) — если можно идти сразу.
func (a *App) termsRequired(ctx context.Context, chatID int64) (string, bool) {
	a.mu.Lock()
	text := ""
	if a.botCfg != nil {
		text = a.botCfg.Contact.TermsText
	}
	a.mu.Unlock()
	if text == "" || a.store == nil {
		return "", false
	}
	u, err := a.store.GetUser(ctx, chatID)
	if err != nil || u == nil {
		// Пользователь ещё не в БД — соглашение тоже не показываем (его сначала
		// зарегистрируют — touch при /start через rememberUser). Безопасно
		// вернуть false; в худшем случае попросим на следующем заходе.
		return "", false
	}
	if u.TermsAcceptedAt != "" {
		return "", false
	}
	return text, true
}

// askTerms показывает текст соглашения и две кнопки.
func (a *App) askTerms(ctx context.Context, chatID int64, text string) {
	lang := a.lang(chatID)
	a.sendKB(ctx, chatID, i18n.T(lang, "terms.intro")+"\n\n"+text, [][]models.InlineKeyboardButton{
		{btn(i18n.T(lang, "terms.btn_accept"), "terms:accept")},
		{btn(i18n.T(lang, "terms.btn_decline"), "terms:decline")},
	})
}

// onTerms принимает решение пользователя по соглашению.
//   - accept → пишем в БД и сразу открываем выбор тарифа.
//   - decline → возвращаем на главную (мягко, без шантажа).
func (a *App) onTerms(ctx context.Context, chatID int64, val, firstName, username string) {
	switch val {
	case "accept":
		if a.store != nil {
			_ = a.store.SetTermsAccepted(ctx, chatID, time.Now().UTC().Format(time.RFC3339))
		}
		a.showPlans(ctx, chatID)
	case "decline":
		isAdmin := chatID == a.cfg.AdminID
		a.enterHome(ctx, chatID, isAdmin, firstName, username)
	}
}
