package app

import (
	"context"
	"errors"

	"remnabot/internal/model"
)

// miniPayURL creates a payment for an external method and returns a URL the
// Mini App can open: a Telegram invoice link (invoice=true → openInvoice) for
// Stars, or a payment-page/redirect URL (openLink) for the others. It reuses
// the SAME invoice-creation cores as the chat flow, so pending-invoice ExtID
// formats are identical and the existing webhooks complete the payment.
func (a *App) miniPayURL(ctx context.Context, tgID int64, months int, method string) (string, bool, error) {
	switch method {
	case model.PayMethodStars:
		link, err := a.starsInvoiceLink(ctx, tgID, months)
		return link, true, err

	case model.PayMethodYooKassa:
		cfg := a.ykConfig()
		pr := a.pricing()
		value := pr.Fiat(model.PayMethodYooKassa, months)
		if !cfg.Enabled || value == "" {
			return "", false, errors.New("оплата картой недоступна")
		}
		returnURL := cfg.ReturnURL
		if returnURL == "" {
			returnURL = "https://t.me"
		}
		currency := pr.Currency
		if len(currency) != 3 {
			currency = "RUB"
		}
		desc := miniDesc(months)
		url, _, err := a.ykCreatePayment(ctx, tgID, months, value, currency, returnURL, desc)
		return url, false, err

	case model.PayMethodCryptoBot:
		cfg := a.cbConfig()
		price := a.pricing().Base[months]
		if !cfg.Enabled || price == "" {
			return "", false, errors.New("оплата криптовалютой недоступна")
		}
		url, _, err := a.cbCreateInvoice(ctx, tgID, months, price)
		return url, false, err

	case model.PayMethodPlatega:
		cfg := a.plConfig()
		pr := a.pricing()
		value := pr.Fiat(model.PayMethodPlatega, months)
		if !cfg.Enabled || value == "" {
			return "", false, errors.New("оплата недоступна")
		}
		returnURL := cfg.ReturnURL
		if returnURL == "" {
			returnURL = "https://t.me"
		}
		url, _, err := a.plCreateTransaction(ctx, tgID, months, parseAmountRub(value), miniDesc(months), returnURL)
		return url, false, err

	case model.PayMethodTribute:
		cfg := a.tributeCfg()
		if !cfg.Enabled || cfg.PayURL == "" {
			return "", false, errors.New("оплата недоступна")
		}
		if a.store != nil {
			_ = a.store.UpsertUser(ctx, tgID)
		}
		return cfg.PayURL, false, nil
	}
	return "", false, errors.New("неизвестный способ оплаты")
}

// miniDesc is a neutral invoice description for Mini App payments.
func miniDesc(months int) string {
	return "VPN " + itoaMonths(months)
}

func itoaMonths(m int) string {
	switch m {
	case 1:
		return "1 мес."
	case 3:
		return "3 мес."
	case 6:
		return "6 мес."
	case 12:
		return "12 мес."
	}
	return "подписка"
}
