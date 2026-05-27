// Package model содержит общие типы конфигурации, разделяемые между слоями
// (storage, remnawave, app), чтобы не плодить циклические импорты.
package model

// Поддерживаемые движки БД.
const (
	DBSQLite   = "sqlite"
	DBPostgres = "postgres"
)

// Режим расположения бота относительно панели.
const (
	ModeLocal  = "local"  // бот в одной docker-сети с панелью, бьём в remnawave:3000 напрямую
	ModeRemote = "remote" // бот на другом сервере, ходим через публичный HTTPS-домен
)

// Способ установки панели (влияет на тип защиты публичного /api).
const (
	InstallDocs   = "docs"   // официальная установка (Caddy + caddy-with-auth)
	InstallEGames = "egames" // скрипт eGames (nginx + защита по куке)
)

// Языки интерфейса.
const (
	LangRU = "ru"
	LangEN = "en"
)

// PanelConfig — параметры подключения к панели Remnawave.
//
// Какие поля заполняются, зависит от Mode и InstallType:
//   - local:            BaseURL подставляется автоматически, Cookie/APIKey не нужны.
//   - remote + egames:  нужен Cookie ("ИМЯ=ЗНАЧЕНИЕ" из nginx.conf панели).
//   - remote + docs:    APIKey (X-API-Key) — только если оператор защитил /api в Caddy.
type PanelConfig struct {
	Mode        string `json:"mode"`
	InstallType string `json:"install_type"`
	BaseURL     string `json:"base_url"`
	APIToken    string `json:"api_token"` // Bearer-токен панели (role API)
	Cookie      string `json:"cookie"`    // "name=value" для eGames nginx, иначе ""
	APIKey      string `json:"api_key"`   // X-API-Key для защищённого Caddy /api, иначе ""
}

// BotConfig — вся конфигурация бота, хранится одной зашифрованной строкой в БД.
type BotConfig struct {
	Installed bool        `json:"installed"`
	Language  string      `json:"language"`
	DBKind    string      `json:"db_kind"`
	Panel     PanelConfig `json:"panel"`
}
