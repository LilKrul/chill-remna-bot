package assets

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed sections/*.jpg
var sectionsFS embed.FS

const (
	SectionWizardWelcome       = "wizard_welcome"
	SectionWizardDBChoose      = "wizard_db_choose"
	SectionWizardDBPostgresUp  = "wizard_db_pg_up"
	SectionWizardLocation      = "wizard_location"
	SectionWizardInstallChoice = "wizard_install_choice"
	SectionWizardToken         = "wizard_token"
	SectionWizardCookie        = "wizard_cookie"
	SectionWizardVerifyOK      = "wizard_verify_ok"

	SectionMainMenu        = "main_menu"
	SectionBuySubscription = "buy_subscription"
	SectionMySubscription  = "my_subscription"
	SectionTrial           = "trial"
	SectionReferral        = "referral"
	SectionPromoCode       = "promo_code"
	SectionAdminStats      = "admin_stats"
)

var SectionImages = map[string]string{

	SectionWizardWelcome:       "https://plus.unsplash.com/premium_photo-1674476932936-80a969879ec2?w=1280&h=640&fit=crop&auto=format&q=80",
	SectionWizardDBChoose:      "https://plus.unsplash.com/premium_photo-1661386261378-8ed99f4e37ba?w=1280&h=640&fit=crop&auto=format&q=80",
	SectionWizardDBPostgresUp:  "https://images.unsplash.com/photo-1775616788028-ce670411dff7?w=1280&h=640&fit=crop&auto=format&q=80",
	SectionWizardLocation:      "https://images.unsplash.com/photo-1778452419724-1f605dc17c46?w=1280&h=640&fit=crop&auto=format&q=80",
	SectionWizardInstallChoice: "https://images.unsplash.com/photo-1518773553398-650c184e0bb3?w=1280&h=640&fit=crop&auto=format&q=80",
	SectionWizardToken:         "https://images.unsplash.com/photo-1608390063578-8dcd6c1995e8?w=1280&h=640&fit=crop&auto=format&q=80",
	SectionWizardCookie:        "https://images.unsplash.com/photo-1497051788611-2c64812349fa?w=1280&h=640&fit=crop&auto=format&q=80",
	SectionWizardVerifyOK:      "https://images.unsplash.com/photo-1767260408878-4566afa38b9c?w=1280&h=640&fit=crop&auto=format&q=80",

	SectionMainMenu:        "https://plus.unsplash.com/premium_photo-1733306489269-449d1e8ae119?w=1280&h=640&fit=crop&auto=format&q=80",
	SectionBuySubscription: "https://images.unsplash.com/photo-1757185389479-6f9c6d02b96d?w=1280&h=640&fit=crop&auto=format&q=80",
	SectionMySubscription:  "https://images.unsplash.com/photo-1744782211816-c5224434614f?w=1280&h=640&fit=crop&auto=format&q=80",
	SectionTrial:           "https://images.unsplash.com/photo-1764385827253-3d0a5eb813fe?w=1280&h=640&fit=crop&auto=format&q=80",
	SectionReferral:        "https://images.unsplash.com/photo-1761075666032-7540b8c58de7?w=1280&h=640&fit=crop&auto=format&q=80",
	SectionPromoCode:       "https://plus.unsplash.com/premium_photo-1681398745480-151fc6addaaf?w=1280&h=640&fit=crop&auto=format&q=80",
	SectionAdminStats:      "https://images.unsplash.com/photo-1745270917233-65e776a47547?w=1280&h=640&fit=crop&auto=format&q=80",
}

type Section struct {
	Key     string
	LabelRU string
	LabelEN string
}

var AllSections = []Section{

	{SectionWizardWelcome, "👋 Приветствие мастера", "👋 Wizard welcome"},
	{SectionWizardDBChoose, "🗄 Шаг: выбор БД", "🗄 Step: DB choice"},
	{SectionWizardDBPostgresUp, "🐘 Шаг: PostgreSQL up", "🐘 Step: PostgreSQL up"},
	{SectionWizardLocation, "📍 Шаг: локально/удалённо", "📍 Step: local/remote"},
	{SectionWizardInstallChoice, "🧩 Шаг: способ установки", "🧩 Step: install type"},
	{SectionWizardToken, "🔑 Шаг: API-токен", "🔑 Step: API token"},
	{SectionWizardCookie, "🍪 Шаг: nginx-кука", "🍪 Step: nginx cookie"},
	{SectionWizardVerifyOK, "✅ Шаг: проверка успешна", "✅ Step: verify OK"},

	{SectionMainMenu, "🏠 Меню «Интерфейс»", "🏠 Menu «Interface»"},
	{SectionBuySubscription, "🛒 Купить / Оплата", "🛒 Buy / Payments menu"},
	{SectionMySubscription, "📦 Мои подписки", "📦 My subscriptions"},
	{SectionTrial, "🎁 Триал", "🎁 Trial"},
	{SectionReferral, "👥 Пользователи / реферал", "👥 Users / referral"},
	{SectionPromoCode, "💸 Платежи / промокод", "💸 Payments / promo"},
	{SectionAdminStats, "⚙️ Управление", "⚙️ Manage"},
}

var userFacingSections = map[string]bool{
	SectionBuySubscription: true,
	SectionMySubscription:  true,
}

func UserSections() []Section {
	out := make([]Section, 0, len(userFacingSections))
	for _, s := range AllSections {
		if userFacingSections[s.Key] {
			out = append(out, s)
		}
	}
	return out
}

func LabelByKey(key, lang string) string {
	for _, sec := range AllSections {
		if sec.Key == key {
			if lang == "en" {
				return sec.LabelEN
			}
			return sec.LabelRU
		}
	}
	return key
}

func URL(section string) string {
	return SectionImages[section]
}

func Bytes(section string) []byte {

	name := "sections/" + strings.TrimSpace(section) + ".jpg"
	data, err := fs.ReadFile(sectionsFS, name)
	if err != nil {
		return nil
	}
	return data
}
