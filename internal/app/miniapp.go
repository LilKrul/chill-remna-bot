package app

import (
	"context"

	"remnabot/internal/model"
	"remnabot/internal/web"
)

// This file implements web.MiniProvider: thin, read-mostly adapters that expose
// the bot's EXISTING data/predicates to the Mini App API. No business logic is
// duplicated here — every value mirrors what the chat bot already computes, so
// the Mini App can never offer an action the bot doesn't have.

// MiniEnabled reports the Mini App feature flag.
func (a *App) MiniEnabled() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.botCfg != nil && a.botCfg.MiniApp.Enabled
}

// MiniBotToken returns the Telegram bot token (used for init-data validation).
func (a *App) MiniBotToken() string { return a.cfg.BotToken }

// MiniMe returns the user's basic profile (balance, language).
func (a *App) MiniMe(ctx context.Context, tgID int64) web.MiniMeDTO {
	dto := web.MiniMeDTO{TgID: tgID, Lang: a.lang(tgID)}
	if a.store != nil {
		if u, _ := a.store.GetUser(ctx, tgID); u != nil {
			dto.BalanceK = u.Balance
		}
	}
	return dto
}

// MiniMenu mirrors navRow: it reports exactly which actions the chat bot would
// offer this user, plus the enabled payment methods and contact links.
func (a *App) MiniMenu(ctx context.Context, tgID int64) web.MiniMenuDTO {
	dto := web.MiniMenuDTO{
		HasSub:         a.userHasSub(ctx, tgID),
		TrialAvailable: a.trialAvailable(ctx, tgID),
		ReferralOn:     a.referralCfg().Enabled,
		SupportURL:     a.supportURL(),
	}
	if dto.HasSub {
		dto.CanRenew = a.renewEligible(ctx, tgID)
	}
	a.mu.Lock()
	if a.botCfg != nil {
		c := a.botCfg
		dto.GroupURL = c.Contact.GroupURL
		if c.Stars.Enabled {
			dto.PayMethods = append(dto.PayMethods, model.PayMethodStars)
		}
		if c.YooKassa.Enabled {
			dto.PayMethods = append(dto.PayMethods, model.PayMethodYooKassa)
		}
		if c.CryptoBot.Enabled {
			dto.PayMethods = append(dto.PayMethods, model.PayMethodCryptoBot)
		}
		if c.Platega.Enabled {
			dto.PayMethods = append(dto.PayMethods, model.PayMethodPlatega)
		}
		if c.Tribute.Enabled {
			dto.PayMethods = append(dto.PayMethods, model.PayMethodTribute)
		}
		if c.P2P.Enabled {
			dto.PayMethods = append(dto.PayMethods, model.PayMethodP2P)
		}
	}
	a.mu.Unlock()
	return dto
}

// MiniSubscription mirrors showMySubs: link, expiry, status, and the read-only
// devices count (only the connected number is sent when no per-user limit).
func (a *App) MiniSubscription(ctx context.Context, tgID int64) web.MiniSubDTO {
	a.mu.Lock()
	panel := a.panel
	a.mu.Unlock()
	var dto web.MiniSubDTO
	if panel == nil {
		return dto
	}
	url, expireAt, status, ok := panel.SubscriptionFull(ctx, tgID)
	if !ok {
		return dto
	}
	dto.Active = true
	dto.Status = status
	dto.SubURL = a.rewriteSub(url)
	dto.ExpireAt = formatExpire(expireAt, a.lang(tgID))
	if info, dok := panel.DevicesByTelegramID(ctx, tgID); dok {
		dto.DevicesOK = true
		dto.DevicesUsed = info.Used
		dto.DeviceLimit = info.Limit
		dto.HasLimit = info.HasLimit
	}
	return dto
}

// MiniPlans mirrors the storefront periods (model.PlanMonths) with base prices
// and included limits from model.Pricing.
func (a *App) MiniPlans(ctx context.Context, tgID int64) web.MiniPlansDTO {
	a.mu.Lock()
	defer a.mu.Unlock()
	var dto web.MiniPlansDTO
	if a.botCfg == nil {
		return dto
	}
	p := a.botCfg.Pricing
	for _, m := range model.PlanMonths {
		price := p.Base[m]
		if price == "" {
			continue
		}
		dto.Plans = append(dto.Plans, web.MiniPlanDTO{
			Months:    m,
			Price:     price,
			Currency:  p.Currency,
			TrafficGB: p.Traffic[m],
			Devices:   p.DeviceLimitFor(m),
		})
	}
	return dto
}
