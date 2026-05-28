package app

import (
	"net/url"
	"strings"
)

// rewriteSubLink подменяет хост в ссылке подписки на overrideDomain,
// сохраняя путь (в нём short-id панели) и схему. Если override пуст,
// либо ссылка пустая, либо хост уже совпадает — возвращает src без изменений.
//
// Примеры:
//
//	src = "https://panel.example.com/sub/aBcD1234"
//	override = "vpn.mybrand.io"
//	→ "https://vpn.mybrand.io/sub/aBcD1234"
//
// Допустимо передавать override со схемой ("https://vpn.x.io") — она
// игнорируется, берётся только host. Полностью невалидный src возвращается
// как есть (не ломаем покупку из-за опечатки в настройке).
func rewriteSubLink(src, override string) string {
	override = strings.TrimSpace(override)
	if override == "" || src == "" {
		return src
	}
	host := extractHost(override)
	if host == "" {
		return src
	}
	u, err := url.Parse(src)
	if err != nil || u.Host == "" {
		return src
	}
	if u.Host == host {
		return src
	}
	u.Host = host
	return u.String()
}

// extractHost вытаскивает чистый host из строки: принимает "vpn.x.io",
// "https://vpn.x.io", "https://vpn.x.io/sub/...".
func extractHost(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "://") {
		// без схемы — это уже host (или host:port).
		// Убираем возможный лишний путь.
		if i := strings.Index(s, "/"); i >= 0 {
			s = s[:i]
		}
		return s
	}
	u, err := url.Parse(s)
	if err != nil {
		return ""
	}
	return u.Host
}

// subOverride возвращает текущий override-домен из конфига (потокобезопасно).
func (a *App) subOverride() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.botCfg == nil {
		return ""
	}
	return a.botCfg.SubscriptionDomain
}

// rewriteSub применяет subOverride() к ссылке.
func (a *App) rewriteSub(src string) string {
	return rewriteSubLink(src, a.subOverride())
}
