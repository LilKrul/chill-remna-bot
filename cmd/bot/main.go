// Команда bot — точка входа Telegram-бота Remnawave.
//
// Окружение (bootstrap-минимум, остальное настраивается мастером в Telegram):
//
//	BOT_TOKEN          — токен бота (обязательно)
//	ADMIN_TELEGRAM_ID  — Telegram ID администратора (обязательно)
//	DATA_DIR           — каталог данных (по умолчанию /data)
//	DB_KIND            — необязательно: sqlite|postgres (иначе спросит мастер)
//	DATABASE_URL       — DSN PostgreSQL (если DB_KIND=postgres)
//	SECRET_KEY         — ключ шифрования секретов (иначе сгенерируется в DATA_DIR)
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"remnabot/internal/app"
	"remnabot/internal/config"
	"remnabot/internal/crypto"

	_ "remnabot/internal/storage/drivers" // регистрация драйверов БД (sqlite, pgx)
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("конфигурация", "err", err)
		os.Exit(1)
	}

	crypter, err := crypto.LoadOrCreate(cfg.SecretKey, cfg.DataDir)
	if err != nil {
		log.Error("ключ шифрования", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a := app.New(cfg, crypter, log)
	if err := a.Bootstrap(ctx); err != nil {
		log.Error("инициализация", "err", err)
		os.Exit(1)
	}

	if err := a.Run(ctx); err != nil {
		log.Error("работа бота", "err", err)
		os.Exit(1)
	}
	log.Info("остановлен")
}
