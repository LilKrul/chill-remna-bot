package app

import (
	"context"
	"testing"

	"remnabot/internal/model"
)

func TestCabinetEmailRegisterLogin(t *testing.T) {
	a, _, fs := newTestApp(t)
	a.store = fs
	a.botCfg = &model.BotConfig{Installed: true}
	a.botCfg.NormalizeCabinet()
	a.botCfg.Cabinet.Enabled = true
	ctx := context.Background()

	id, err := a.CabinetEmailRegister(ctx, "User@Example.com", "secret12")
	if err != nil || id >= 0 {
		t.Fatalf("register: id=%d err=%v (id must be negative synthetic)", id, err)
	}
	if u, _ := fs.GetUser(ctx, id); u == nil {
		t.Fatal("local user not created for web account")
	}
	if _, err := a.CabinetEmailRegister(ctx, "user@example.com", "secret12"); err == nil {
		t.Fatal("duplicate email must be rejected (case-insensitive)")
	}
	if lid, err := a.CabinetEmailLogin(ctx, "user@example.com", "secret12"); err != nil || lid != id {
		t.Fatalf("login: lid=%d err=%v", lid, err)
	}
	if _, err := a.CabinetEmailLogin(ctx, "user@example.com", "wrong"); err == nil {
		t.Fatal("wrong password must be rejected")
	}
	if _, err := a.CabinetEmailRegister(ctx, "a@b.io", "123"); err == nil {
		t.Fatal("short password must be rejected")
	}
}
