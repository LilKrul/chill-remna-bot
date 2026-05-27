// Package config читает параметры окружения, необходимые до подключения к БД.
//
// Философия проекта: env содержит только bootstrap-минимум, вся остальная
// настройка делается мастером прямо в Telegram и хранится в БД.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	BotToken string // токен Telegram-бота (обязательно)
	AdminID  int64  // Telegram ID администратора, которому доступен мастер (обязательно)
	DataDir  string // каталог для файла SQLite, bootstrap.json и secret.key

	// Необязательное: если заданы в env (например, при запуске compose-профиля
	// postgres), мастер не будет спрашивать движок/DSN, а возьмёт отсюда.
	DBKind      string // "sqlite" | "postgres" | ""
	DatabaseURL string // DSN PostgreSQL, если DBKind == postgres
	SecretKey   string // ключ шифрования секретов; если пуст — сгенерируем в DataDir
}

func Load() (*Config, error) {
	c := &Config{
		BotToken:    strings.TrimSpace(os.Getenv("BOT_TOKEN")),
		DataDir:     envOr("DATA_DIR", "/data"),
		DBKind:      strings.TrimSpace(os.Getenv("DB_KIND")),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		SecretKey:   os.Getenv("SECRET_KEY"),
	}
	if c.BotToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN не задан")
	}
	rawAdmin := strings.TrimSpace(os.Getenv("ADMIN_TELEGRAM_ID"))
	if rawAdmin == "" {
		return nil, fmt.Errorf("ADMIN_TELEGRAM_ID не задан")
	}
	id, err := strconv.ParseInt(rawAdmin, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ADMIN_TELEGRAM_ID должен быть числом: %w", err)
	}
	c.AdminID = id
	return c, nil
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
