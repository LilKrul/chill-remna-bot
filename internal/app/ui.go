package app

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/go-telegram/bot/models"

	"remnabot/internal/i18n"
	"remnabot/internal/model"
)

//go:embed banner_default.jpg
var defaultBanner []byte

// botEmojis — эмодзи, используемые в сообщениях бота (без чисто служебных
// настроечных), и где они применяются. Используется в /emoji.
var botEmojis = []struct{ E, Use string }{
	{"👋", "приветствие"}, {"✅", "успех / подтверждение"}, {"❌", "ошибка / отказ"},
	{"⚠️", "предупреждение"}, {"⏳", "ожидание"}, {"🕒", "на проверке"},
	{"💳", "оплата"}, {"📦", "тариф"}, {"💸", "платёж"}, {"🔒", "доступ"},
	{"🔔", "уведомление"}, {"📸", "скриншот"}, {"📊", "статус"}, {"🌐", "удалённо"},
	{"🏠", "локально"}, {"🔑", "токен"}, {"🍪", "кука"},
}

func (a *App) botLang() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.botCfg != nil && a.botCfg.Language != "" {
		return a.botCfg.Language
	}
	return i18n.Fallback
}

// replyKeyboard — постоянная клавиатура снизу: свой набор для админа и юзера.
func (a *App) replyKeyboard(isAdmin bool) models.ReplyMarkup {
	lang := a.botLang()
	if !isAdmin {
		return models.ReplyKeyboardMarkup{
			Keyboard:       [][]models.KeyboardButton{{{Text: i18n.T(lang, "btn.buy")}}},
			ResizeKeyboard: true,
			IsPersistent:   true,
		}
	}
	return models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{{Text: i18n.T(lang, "btn.buy")}, {Text: i18n.T(lang, "btn.status")}},
			{{Text: i18n.T(lang, "btn.p2p")}, {Text: i18n.T(lang, "btn.emoji")}},
			{{Text: i18n.T(lang, "btn.banner")}, {Text: i18n.T(lang, "btn.update")}},
		},
		ResizeKeyboard: true,
		IsPersistent:   true,
	}
}

// handleReplyButton обрабатывает нажатия постоянной клавиатуры (текст кнопки).
func (a *App) handleReplyButton(ctx context.Context, chatID int64, text string, isAdmin bool) bool {
	lang := a.botLang()
	switch text {
	case i18n.T(lang, "btn.buy"):
		a.showPlans(ctx, chatID)
		return true
	}
	if !isAdmin {
		return false
	}
	switch text {
	case i18n.T(lang, "btn.status"):
		a.handleStatus(ctx, chatID)
		return true
	case i18n.T(lang, "btn.p2p"):
		a.showP2PAdmin(ctx, chatID)
		return true
	case i18n.T(lang, "btn.emoji"):
		a.showEmojiGrid(ctx, chatID)
		return true
	case i18n.T(lang, "btn.banner"):
		a.showWelcomeAdmin(ctx, chatID)
		return true
	case i18n.T(lang, "btn.update"):
		a.handleUpdate(ctx, chatID)
		return true
	}
	return false
}

// --- стартовый баннер ---

func (a *App) welcomeContent() (models.InputFile, string, []models.MessageEntity) {
	a.mu.Lock()
	var w model.WelcomeConfig
	lang := i18n.Fallback
	if a.botCfg != nil {
		w = a.botCfg.Welcome
		if a.botCfg.Language != "" {
			lang = a.botCfg.Language
		}
	}
	a.mu.Unlock()

	var photo models.InputFile
	switch {
	case w.ImageFileID != "":
		photo = &models.InputFileString{Data: w.ImageFileID}
	case w.ImageURL != "":
		photo = &models.InputFileString{Data: w.ImageURL}
	default:
		photo = &models.InputFileUpload{Filename: "welcome.jpg", Data: bytes.NewReader(defaultBanner)}
	}

	caption := w.Text
	var ents []models.MessageEntity
	if caption == "" {
		caption = i18n.T(lang, "menu.welcome")
	} else if len(w.Entities) > 0 {
		_ = json.Unmarshal(w.Entities, &ents)
	}
	return photo, caption, ents
}

func (a *App) showMenu(ctx context.Context, chatID int64, isAdmin bool) {
	photo, caption, ents := a.welcomeContent()
	if len(ents) == 0 {
		caption = a.applyPremium(caption)
	}
	a.msg.SendBanner(ctx, chatID, photo, caption, ents, a.replyKeyboard(isAdmin))
}

// --- админ: баннер ---

func (a *App) showWelcomeAdmin(ctx context.Context, chatID int64) {
	lang := a.lang(chatID)
	a.sendKB(ctx, chatID, i18n.T(lang, "welcome.title"), [][]models.InlineKeyboardButton{{
		btn(i18n.T(lang, "welcome.btn_image"), "wel:img"),
		btn(i18n.T(lang, "welcome.btn_text"), "wel:txt"),
	}})
}

func (a *App) onWelcome(ctx context.Context, chatID int64, val string) {
	ui := a.getUI(chatID)
	switch val {
	case "img":
		ui.welcomeAwait = "img"
		a.send(ctx, chatID, i18n.T(a.lang(chatID), "welcome.ask_image"))
	case "txt":
		ui.welcomeAwait = "txt"
		a.send(ctx, chatID, i18n.T(a.lang(chatID), "welcome.ask_text"))
	}
}

func (a *App) setWelcomeImageURL(ctx context.Context, chatID int64, url string) {
	a.getUI(chatID).welcomeAwait = ""
	a.mu.Lock()
	if a.botCfg != nil {
		a.botCfg.Welcome.ImageURL = strings.TrimSpace(url)
		a.botCfg.Welcome.ImageFileID = ""
	}
	a.mu.Unlock()
	_ = a.saveBotConfig(ctx)
	a.send(ctx, chatID, i18n.T(a.lang(chatID), "welcome.saved"))
}

func (a *App) setWelcomeImageFile(ctx context.Context, chatID int64, fileID string) {
	a.getUI(chatID).welcomeAwait = ""
	a.mu.Lock()
	if a.botCfg != nil {
		a.botCfg.Welcome.ImageFileID = fileID
		a.botCfg.Welcome.ImageURL = ""
	}
	a.mu.Unlock()
	_ = a.saveBotConfig(ctx)
	a.send(ctx, chatID, i18n.T(a.lang(chatID), "welcome.saved"))
}

func (a *App) setWelcomeText(ctx context.Context, chatID int64, m *models.Message) {
	a.getUI(chatID).welcomeAwait = ""
	ents, _ := json.Marshal(m.Entities)
	a.mu.Lock()
	if a.botCfg != nil {
		a.botCfg.Welcome.Text = m.Text
		a.botCfg.Welcome.Entities = ents
	}
	a.mu.Unlock()
	_ = a.saveBotConfig(ctx)
	a.send(ctx, chatID, i18n.T(a.lang(chatID), "welcome.saved"))
}

// --- админ: эмодзи (грид) ---

func (a *App) showEmojiGrid(ctx context.Context, chatID int64) {
	lang := a.lang(chatID)
	m := a.premiumMap()
	var sb strings.Builder
	sb.WriteString(i18n.T(lang, "emoji.title"))
	for _, e := range botEmojis {
		sb.WriteString("\n" + e.E + " — " + e.Use)
	}
	var rows [][]models.InlineKeyboardButton
	var row []models.InlineKeyboardButton
	for _, e := range botEmojis {
		label := e.E
		if _, ok := m[e.E]; ok {
			label = e.E + " ✅"
		}
		row = append(row, btn(label, "emo:set:"+e.E))
		if len(row) == 3 {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	rows = append(rows, []models.InlineKeyboardButton{btn(i18n.T(lang, "emoji.btn_done"), "emo:done")})
	a.sendKB(ctx, chatID, sb.String(), rows)
}

func (a *App) onEmoji(ctx context.Context, chatID int64, val string) {
	action, arg, _ := strings.Cut(val, ":")
	switch action {
	case "set":
		a.getUI(chatID).awaitEmojiFor = arg
		a.send(ctx, chatID, i18n.T(a.lang(chatID), "emoji.ask_one", arg))
	case "done":
		a.getUI(chatID).awaitEmojiFor = ""
		a.send(ctx, chatID, i18n.T(a.lang(chatID), "admin.done"))
	}
}

func (a *App) setEmojiFor(ctx context.Context, chatID int64, m *models.Message) {
	ui := a.getUI(chatID)
	target := ui.awaitEmojiFor
	ui.awaitEmojiFor = ""
	var id string
	for _, e := range m.Entities {
		if e.Type == models.MessageEntityTypeCustomEmoji && e.CustomEmojiID != "" {
			id = e.CustomEmojiID
			break
		}
	}
	if id == "" {
		a.send(ctx, chatID, i18n.T(a.lang(chatID), "emoji.none_in_msg"))
		a.showEmojiGrid(ctx, chatID)
		return
	}
	a.mu.Lock()
	if a.botCfg != nil {
		if a.botCfg.PremiumEmoji == nil {
			a.botCfg.PremiumEmoji = map[string]string{}
		}
		a.botCfg.PremiumEmoji[target] = id
	}
	a.mu.Unlock()
	_ = a.saveBotConfig(ctx)
	a.send(ctx, chatID, i18n.T(a.lang(chatID), "emoji.set_ok", target))
	a.showEmojiGrid(ctx, chatID)
}
