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
