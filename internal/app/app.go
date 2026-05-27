// Package app связывает воедино конфиг, хранилище, клиент панели и
// Telegram-бота, и реализует мастер первичной установки (FSM).
//
// На текущем этапе реализована ТОЛЬКО установка бота (выбор БД, подключение
// панели с ветвлением по способу установки) + команда /status. Магазин,
// платежи и админ-функции добавляются на следующих фазах.
package app

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"remnabot/internal/config"
	"remnabot/internal/crypto"
	"remnabot/internal/i18n"
	"remnabot/internal/model"
	"remnabot/internal/remnawave"
	"remnabot/internal/storage"
)

type App struct {
	cfg     *config.Config
	crypter *crypto.Crypter
	log     *slog.Logger
	b       *bot.Bot

	mu     sync.Mutex
	store  storage.Storage   // nil, пока БД не выбрана в мастере
	botCfg *model.BotConfig  // nil, пока бот не установлен
	panel  *remnawave.Client // nil, пока бот не установлен
	wiz    map[int64]*wizard // состояние мастера по chatID
}

func New(cfg *config.Config, crypter *crypto.Crypter, log *slog.Logger) *App {
	return &App{cfg: cfg, crypter: crypter, log: log, wiz: map[int64]*wizard{}}
}

// Bootstrap при старте подхватывает ранее выбранную БД и конфиг (если бот
// уже был установлен), чтобы продолжить работу без повторной установки.
func (a *App) Bootstrap(ctx context.Context) error {
	bs, err := storage.LoadBootstrap(a.cfg.DataDir)
	if err != nil {
		return err
	}
	if bs == nil {
		// Первый запуск: БД ещё не выбрана. Если движок задан в env — можно
		// открыть заранее, иначе ждём выбора в мастере.
		if a.cfg.DBKind != "" {
			dsn := a.dsnForEnv(a.cfg.DBKind)
			if err := a.openStore(a.cfg.DBKind, dsn); err != nil {
				return err
			}
		}
		return a.loadConfigIfStore(ctx)
	}
	if err := a.openStore(bs.DBKind, bs.DSN); err != nil {
		return err
	}
	return a.loadConfigIfStore(ctx)
}

func (a *App) dsnForEnv(kind string) string {
	if kind == model.DBPostgres {
		return a.cfg.DatabaseURL
	}
	return filepath.Join(a.cfg.DataDir, "bot.db")
}

func (a *App) loadConfigIfStore(ctx context.Context) error {
	if a.store == nil {
		return nil
	}
	cfg, ok, err := a.store.LoadConfig(ctx)
	if err != nil {
		return err
	}
	if ok && cfg.Installed {
		a.botCfg = cfg
		a.panel = remnawave.New(cfg.Panel)
		a.log.Info("конфигурация загружена, бот установлен", "db", a.store.Kind())
	}
	return nil
}

// openStore открывает БД, прогоняет миграции, запоминает выбор в bootstrap.json.
func (a *App) openStore(kind, dsn string) error {
	st, err := storage.Open(kind, dsn, a.crypter)
	if err != nil {
		return err
	}
	if err := st.Migrate(context.Background()); err != nil {
		_ = st.Close()
		return err
	}
	if a.store != nil {
		_ = a.store.Close()
	}
	a.store = st
	return storage.SaveBootstrap(a.cfg.DataDir, &storage.Bootstrap{DBKind: kind, DSN: dsn})
}

// Run создаёт бота и запускает long polling до отмены контекста.
func (a *App) Run(ctx context.Context) error {
	b, err := bot.New(a.cfg.BotToken, bot.WithDefaultHandler(a.handle))
	if err != nil {
		return err
	}
	a.b = b
	a.log.Info("бот запущен")
	b.Start(ctx)
	return nil
}

func (a *App) installed() bool {
	return a.botCfg != nil && a.botCfg.Installed
}

// handle — единая точка входа для всех апдейтов Telegram.
func (a *App) handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	switch {
	case update.CallbackQuery != nil:
		a.handleCallback(ctx, update.CallbackQuery)
	case update.Message != nil && update.Message.Text != "":
		a.handleMessage(ctx, update.Message)
	}
}

func (a *App) handleMessage(ctx context.Context, m *models.Message) {
	chatID := m.Chat.ID
	userID := int64(0)
	if m.From != nil {
		userID = m.From.ID
	}
	text := strings.TrimSpace(m.Text)

	// Команды.
	switch {
	case strings.HasPrefix(text, "/start"), strings.HasPrefix(text, "/setup"):
		if userID != a.cfg.AdminID {
			a.send(ctx, chatID, i18n.T(i18n.Fallback, "setup.not_admin"))
			return
		}
		if a.installed() && strings.HasPrefix(text, "/start") {
			a.send(ctx, chatID, i18n.T(a.botCfg.Language, "installed.hint"))
			return
		}
		a.startWizard(ctx, chatID)
		return
	case strings.HasPrefix(text, "/status"):
		a.handleStatus(ctx, chatID)
		return
	}

	// Текстовый ввод в рамках мастера.
	if userID == a.cfg.AdminID {
		a.handleWizardText(ctx, chatID, text)
	}
}

func (a *App) handleStatus(ctx context.Context, chatID int64) {
	a.mu.Lock()
	installed := a.installed()
	panel := a.panel
	var dbKind, mode string
	if installed {
		dbKind = a.botCfg.DBKind
		mode = a.botCfg.Panel.Mode
	}
	lang := a.lang(chatID)
	a.mu.Unlock()

	if !installed || panel == nil {
		a.send(ctx, chatID, i18n.T(lang, "installed.hint"))
		return
	}
	count, err := panel.SystemStats(ctx)
	if err != nil {
		a.send(ctx, chatID, i18n.T(lang, "status.fail", err.Error()))
		return
	}
	a.send(ctx, chatID, i18n.T(lang, "status.line", count, dbKind, mode))
}

// --- утилиты отправки ---

func (a *App) send(ctx context.Context, chatID int64, text string) {
	_, err := a.b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		a.log.Error("send message", "err", err)
	}
}

func (a *App) sendKB(ctx context.Context, chatID int64, text string, rows [][]models.InlineKeyboardButton) {
	_, err := a.b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: rows},
	})
	if err != nil {
		a.log.Error("send keyboard", "err", err)
	}
}

func btn(text, data string) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{Text: text, CallbackData: data}
}

// lang возвращает язык для chatID: из активного мастера, иначе из конфига, иначе fallback.
func (a *App) lang(chatID int64) string {
	if w, ok := a.wiz[chatID]; ok && w.cfg.Language != "" {
		return w.cfg.Language
	}
	if a.botCfg != nil && a.botCfg.Language != "" {
		return a.botCfg.Language
	}
	return i18n.Fallback
}
