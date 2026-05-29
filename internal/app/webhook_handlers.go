package app

import (
	"context"
	"errors"
	"fmt"
)

func (a *App) Healthy(_ context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.store == nil {
		return errors.New("storage not initialised")
	}
	if a.botCfg == nil || !a.botCfg.Installed {
		return errors.New("bot not installed")
	}
	return nil
}

func (a *App) WebhookConfig() (addr string, enabled bool, publicURL string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.botCfg == nil {
		return ":8080", false, ""
	}
	addr = a.botCfg.Webhook.ListenAddr
	if addr == "" {
		addr = ":8080"
	}
	return addr, a.botCfg.Webhook.Enabled, a.botCfg.Webhook.PublicBaseURL
}

func (a *App) PublicWebhookURL(path string) string {
	a.mu.Lock()
	base := ""
	if a.botCfg != nil {
		base = a.botCfg.Webhook.PublicBaseURL
	}
	a.mu.Unlock()
	if base == "" {
		return ""
	}
	return fmt.Sprintf("%s%s", base, path)
}
