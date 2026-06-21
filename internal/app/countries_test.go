package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"remnabot/internal/model"
	"remnabot/internal/remnawave"
)

func TestSplitFlag(t *testing.T) {
	cases := []struct{ in, flag, name string }{
		{"🇩🇪 Германия", "🇩🇪", "Германия"},
		{"🇳🇱 Netherlands #1", "🇳🇱", "Netherlands"},
		{"  🇺🇸 USA-2 ", "🇺🇸", "USA"},
		{"Premium 🇫🇷 France (3)", "🇫🇷", "Premium France"},
		{"no flag here", "", ""},
	}
	for _, c := range cases {
		f, n := splitFlag(c.in)
		if f != c.flag || n != c.name {
			t.Fatalf("splitFlag(%q)=%q,%q want %q,%q", c.in, f, n, c.flag, c.name)
		}
	}
}

func TestPlanCountries_DedupAndFilter(t *testing.T) {
	a := &App{ui: map[int64]*uiState{}}
	a.botCfg = &model.BotConfig{
		Plan: model.SubscriptionPlan{ActiveInternalSquads: []string{"sq1"}},
	}
	a.botCfg.NormalizePricing()
	a.infraCache = &infraCacheEntry{
		fetchedAt: time.Now(),
		squads: []remnawave.SquadFull{
			{UUID: "sq1", InboundsCount: 3, InboundUUIDs: []string{"ib1", "ib2"}},
			{UUID: "sq2", InboundsCount: 1, InboundUUIDs: []string{"ib9"}},
		},
		hosts: []remnawave.Host{
			{Remark: "🇩🇪 Германия #1", InboundUUID: "ib1"},
			{Remark: "🇩🇪 Германия #2", InboundUUID: "ib2"},
			{Remark: "🇳🇱 Нидерланды", InboundUUID: "ib1"},
			{Remark: "🇺🇸 USA", InboundUUID: "ib1", Hidden: true},
			{Remark: "🇫🇷 France", InboundUUID: "ib1", ExcludedSquads: []string{"sq1"}},
			{Remark: "🇯🇵 Japan", InboundUUID: "ib9"},
		},
	}
	cs, inb := a.planCountries(context.Background(), 1)
	if got := strings.Join(cs, "|"); got != "🇩🇪 Германия|🇳🇱 Нидерланды" {
		t.Fatalf("countries=%q want DE|NL deduped", got)
	}
	if inb != 3 {
		t.Fatalf("inbounds=%d want 3", inb)
	}
}
