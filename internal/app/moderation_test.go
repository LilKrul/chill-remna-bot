package app

import (
	"context"
	"testing"
)

func TestDenyAccess_BlockAndWhitelist(t *testing.T) {
	a, fs := refTestApp(t)
	ctx := context.Background()

	if a.denyAccess(ctx, 300, false) {
		t.Fatal("в обычном режиме новый юзер не должен блокироваться")
	}
	_ = fs.UpsertUser(ctx, 300)
	_ = fs.SetBlocked(ctx, 300, true)
	if !a.denyAccess(ctx, 300, false) {
		t.Fatal("забаненный должен быть заблокирован")
	}
	if a.denyAccess(ctx, 100, true) {
		t.Fatal("админ не блокируется никогда")
	}

	a.botCfg.WhitelistMode = true
	_ = fs.UpsertUser(ctx, 400)
	if !a.denyAccess(ctx, 400, false) {
		t.Fatal("в режиме вайтлиста не-вайтлистнутый не имеет доступа")
	}
	_ = fs.SetWhitelisted(ctx, 400, true)
	if a.denyAccess(ctx, 400, false) {
		t.Fatal("вайтлистнутый должен иметь доступ")
	}
}

func TestDenyAccess_PrefilledWhitelistID(t *testing.T) {
	a, fs := refTestApp(t)
	ctx := context.Background()

	a.botCfg.WhitelistMode = true

	// ID добавлен в вайтлист заранее, юзер ещё не зарегистрирован в боте
	_ = fs.AddWhitelistID(ctx, 777)
	if a.denyAccess(ctx, 777, false) {
		t.Fatal("предзаполненный ID должен иметь доступ даже без регистрации")
	}

	// после удаления из вайтлиста доступ пропадает
	_ = fs.RemoveWhitelistID(ctx, 777)
	if !a.denyAccess(ctx, 777, false) {
		t.Fatal("после удаления из вайтлиста доступа быть не должно")
	}

	ids, _ := fs.ListWhitelistIDs(ctx)
	if len(ids) != 0 {
		t.Fatalf("список вайтлиста должен быть пуст, получили %v", ids)
	}
}
