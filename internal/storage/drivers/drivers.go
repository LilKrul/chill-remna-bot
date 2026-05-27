// Package drivers регистрирует драйверы баз данных (database/sql).
// Импортируется только в точке сборки (cmd/bot) и в тестах storage, чтобы сам
// пакет storage оставался лёгким и не тянул modernc/sqlite в зависимые пакеты.
package drivers

import (
	_ "github.com/jackc/pgx/v5/stdlib" // драйвер "pgx"
	_ "modernc.org/sqlite"             // драйвер "sqlite" (чистый Go, без CGO)
)
