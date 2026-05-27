package storage

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*/*.sql
var migrationsFS embed.FS

// runMigrations применяет ещё не применённые .sql из migrations/<dialect>.
//
// Сознательно простой раннер (одна инструкция на файл), достаточный для текущей
// схемы. По мере роста схемы заменяется на goose без изменения вызывающего кода.
func runMigrations(ctx context.Context, b *base, dialect string) error {
	if _, err := b.db.ExecContext(ctx,
		"CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY)"); err != nil {
		return fmt.Errorf("создание schema_migrations: %w", err)
	}

	dir := "migrations/" + dialect
	entries, err := fs.ReadDir(migrationsFS, dir)
	if err != nil {
		return fmt.Errorf("чтение миграций %s: %w", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		version, err := strconv.Atoi(strings.SplitN(name, "_", 2)[0])
		if err != nil {
			return fmt.Errorf("имя миграции %q должно начинаться с номера: %w", name, err)
		}
		var exists int
		if err := b.db.QueryRowContext(ctx,
			"SELECT COUNT(1) FROM schema_migrations WHERE version = "+b.ph(1), version).
			Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}
		stmt, err := migrationsFS.ReadFile(dir + "/" + name)
		if err != nil {
			return err
		}
		if _, err := b.db.ExecContext(ctx, string(stmt)); err != nil {
			return fmt.Errorf("миграция %s: %w", name, err)
		}
		if _, err := b.db.ExecContext(ctx,
			"INSERT INTO schema_migrations (version) VALUES ("+b.ph(1)+")", version); err != nil {
			return err
		}
	}
	return nil
}
