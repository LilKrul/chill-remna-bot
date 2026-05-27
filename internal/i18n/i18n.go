// Package i18n — простая файловая локализация RU/EN.
//
// Архитектура заложена под мультиязычность: добавление языка — это новая карта
// строк, вызывающий код всегда обращается через T(lang, key, args...).
package i18n

import "fmt"

const Fallback = "ru"

var bundles = map[string]map[string]string{
	"ru": ru,
	"en": en,
}

// Supported возвращает список доступных языков (для кнопок выбора).
func Supported() []string { return []string{"ru", "en"} }

// T возвращает локализованную строку. Если ключа нет в выбранном языке —
// пробует fallback, затем возвращает сам ключ (чтобы пропуск был заметен).
func T(lang, key string, args ...any) string {
	tmpl := lookup(lang, key)
	if tmpl == "" {
		tmpl = lookup(Fallback, key)
	}
	if tmpl == "" {
		return key
	}
	if len(args) == 0 {
		return tmpl
	}
	return fmt.Sprintf(tmpl, args...)
}

func lookup(lang, key string) string {
	if b, ok := bundles[lang]; ok {
		if v, ok := b[key]; ok {
			return v
		}
	}
	return ""
}
