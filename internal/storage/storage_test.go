package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"remnabot/internal/crypto"
	"remnabot/internal/model"

	_ "remnabot/internal/storage/drivers"
)

func testCrypter(t *testing.T) *crypto.Crypter {
	t.Helper()
	c, err := crypto.NewFromKeyMaterial([]byte("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func openSQLiteTest(t *testing.T) Storage {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	st, err := Open(model.DBSQLite, path, testCrypter(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func sampleConfig() *model.BotConfig {
	return &model.BotConfig{
		Installed: true, Language: "ru", DBKind: "sqlite",
		Panel: model.PanelConfig{
			Mode: "remote", InstallType: "egames", BaseURL: "https://p",
			APIToken: "secret-token", Cookie: "AbCdEfGh=IjKlMnOp",
		},
	}
}

func TestSQLiteContract(t *testing.T) {
	ctx := context.Background()
	st := openSQLiteTest(t)

	if _, ok, err := st.LoadConfig(ctx); err != nil || ok {
		t.Fatalf("на пустой БД: ok=%v err=%v", ok, err)
	}
	want := sampleConfig()
	if err := st.SaveConfig(ctx, want); err != nil {
		t.Fatal(err)
	}
	got, ok, err := st.LoadConfig(ctx)
	if err != nil || !ok {
		t.Fatalf("load после save: ok=%v err=%v", ok, err)
	}
	if got.Panel.APIToken != want.Panel.APIToken || got.Language != want.Language || got.Panel.Cookie != want.Panel.Cookie {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
	want.Language = "en"
	if err := st.SaveConfig(ctx, want); err != nil {
		t.Fatal(err)
	}
	got, _, _ = st.LoadConfig(ctx)
	if got.Language != "en" {
		t.Fatalf("upsert не сработал: %q", got.Language)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	st := openSQLiteTest(t)
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestTransferSQLiteToSQLite(t *testing.T) {
	ctx := context.Background()
	src := openSQLiteTest(t)
	dst := openSQLiteTest(t)
	cfg := sampleConfig()
	if err := src.SaveConfig(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if err := Transfer(ctx, src, dst); err != nil {
		t.Fatal(err)
	}
	got, ok, err := dst.LoadConfig(ctx)
	if err != nil || !ok {
		t.Fatalf("load из dst: ok=%v err=%v", ok, err)
	}
	if got.Panel.APIToken != cfg.Panel.APIToken {
		t.Fatal("Transfer потерял данные")
	}
}

// TestPostgresContract запускается, только если задан TEST_POSTGRES_DSN
// (в CI поднимается через сервис postgres). Прогоняет тот же контракт против PG.
func TestPostgresContract(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN не задан")
	}
	ctx := context.Background()
	st, err := Open(model.DBPostgres, dsn, testCrypter(t))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	cfg := sampleConfig()
	if err := st.SaveConfig(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	got, ok, err := st.LoadConfig(ctx)
	if err != nil || !ok || got.Panel.APIToken != cfg.Panel.APIToken {
		t.Fatalf("PG roundtrip провален: ok=%v err=%v", ok, err)
	}
}
