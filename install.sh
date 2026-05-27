#!/usr/bin/env bash
# chill-remna-bot — помощник установки.
# Проверяет окружение, готовит .env и сеть, собирает и запускает бот.
# Дальнейшая настройка (БД, подключение панели) — в Telegram: напишите боту /start.
set -euo pipefail

cd "$(dirname "$0")"

echo "==> Проверяю Docker…"
command -v docker >/dev/null 2>&1 || { echo "Docker не найден. Установите Docker и повторите."; exit 1; }
docker compose version >/dev/null 2>&1 || { echo "Не найден Docker Compose v2 (docker compose). Установите и повторите."; exit 1; }

# --- .env ---
if [ ! -f .env ]; then
  echo "==> .env не найден, создаю из .env.example"
  cp .env.example .env
fi

read_if_empty() {
  # read_if_empty KEY "Подсказка"
  local key="$1" prompt="$2" current
  current="$(grep -E "^${key}=" .env | head -1 | cut -d= -f2- || true)"
  if [ -z "${current}" ]; then
    read -rp "${prompt}: " val
    # экранируем символы для sed
    val="$(printf '%s' "$val" | sed -e 's/[\/&]/\\&/g')"
    if grep -qE "^${key}=" .env; then
      sed -i "s/^${key}=.*/${key}=${val}/" .env
    else
      echo "${key}=${val}" >> .env
    fi
  fi
}

echo "==> Заполняю обязательные параметры (если ещё не заданы)"
read_if_empty BOT_TOKEN "Токен бота от @BotFather"
read_if_empty ADMIN_TELEGRAM_ID "Ваш Telegram ID (число, узнать у @userinfobot)"

# --- сеть ---
echo "==> Создаю общую docker-сеть remnawave-network (если ещё нет)"
docker network create remnawave-network 2>/dev/null || true

# --- выбор БД ---
PROFILE="${1:-}"
if [ -z "${PROFILE}" ]; then
  echo "Выберите базу данных:"
  echo "  1) SQLite       (по умолчанию, для старта)"
  echo "  2) PostgreSQL 17 (для нагруженных проектов)"
  read -rp "Вариант [1/2]: " choice
  case "${choice}" in
    2) PROFILE="postgres" ;;
    *) PROFILE="sqlite" ;;
  esac
fi

echo "==> Собираю и запускаю профиль: ${PROFILE}"
docker compose --profile "${PROFILE}" up -d --build

echo
echo "✅ Готово. Бот запущен (профиль ${PROFILE})."
echo "   Логи:    docker compose --profile ${PROFILE} logs -f"
echo "   Дальше:  откройте чат с ботом в Telegram и отправьте /start"
