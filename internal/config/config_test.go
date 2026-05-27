package config

import "testing"

func TestLoadValid(t *testing.T) {
	t.Setenv("BOT_TOKEN", "tok")
	t.Setenv("ADMIN_TELEGRAM_ID", "12345")
	t.Setenv("DATA_DIR", "/data")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.BotToken != "tok" || c.AdminID != 12345 || c.DataDir != "/data" {
		t.Fatalf("неожиданный конфиг: %+v", c)
	}
}

func TestLoadMissingToken(t *testing.T) {
	t.Setenv("BOT_TOKEN", "")
	t.Setenv("ADMIN_TELEGRAM_ID", "1")
	if _, err := Load(); err == nil {
		t.Fatal("ожидалась ошибка при пустом BOT_TOKEN")
	}
}

func TestLoadBadAdminID(t *testing.T) {
	t.Setenv("BOT_TOKEN", "tok")
	t.Setenv("ADMIN_TELEGRAM_ID", "not-a-number")
	if _, err := Load(); err == nil {
		t.Fatal("ожидалась ошибка при нечисловом ADMIN_TELEGRAM_ID")
	}
}

func TestLoadDefaultDataDir(t *testing.T) {
	t.Setenv("BOT_TOKEN", "tok")
	t.Setenv("ADMIN_TELEGRAM_ID", "1")
	t.Setenv("DATA_DIR", "")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.DataDir != "/data" {
		t.Fatalf("DATA_DIR по умолчанию = %q", c.DataDir)
	}
}
