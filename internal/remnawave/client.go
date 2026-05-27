// Package remnawave — минимальный клиент REST API панели Remnawave.
//
// На этапе установки нужны только проверка связи и базовая статистика;
// методы для юзеров/подписок/платежей добавляются на следующих фазах.
package remnawave

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"remnabot/internal/model"
)

// LocalBaseURL — адрес панели внутри общей docker-сети (минуя reverse-proxy).
const LocalBaseURL = "http://remnawave:3000"

type Client struct {
	base   string
	token  string
	cookie string // "name=value" для eGames(nginx), иначе ""
	apiKey string // X-API-Key для защищённого Caddy, иначе ""
	local  bool
	http   *http.Client
}

func New(cfg model.PanelConfig) *Client {
	base := strings.TrimRight(cfg.BaseURL, "/")
	if cfg.Mode == model.ModeLocal {
		base = LocalBaseURL
	}
	return &Client{
		base:   base,
		token:  cfg.APIToken,
		cookie: strings.TrimSpace(cfg.Cookie),
		apiKey: strings.TrimSpace(cfg.APIKey),
		local:  cfg.Mode == model.ModeLocal,
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	if c.local {
		// ProxyCheckGuard панели в проде рвёт сокет без этих заголовков,
		// когда обращаемся к :3000 напрямую, минуя reverse-proxy.
		req.Header.Set("X-Forwarded-For", "127.0.0.1")
		req.Header.Set("X-Forwarded-Proto", "https")
	}
	if c.cookie != "" {
		// eGames(nginx): без этой куки nginx отдаёт 444 ещё до панели.
		req.Header.Set("Cookie", c.cookie)
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	return req, nil
}

// Health проверяет доступность панели: GET /api/system/health.
func (c *Client) Health(ctx context.Context) error {
	req, err := c.newRequest(ctx, http.MethodGet, "/api/system/health")
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("нет связи с панелью: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return classifyHTTP(resp)
	}
	return nil
}

// SystemStats возвращает счётчик пользователей панели (GET /api/system/stats).
func (c *Client) SystemStats(ctx context.Context) (int, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/api/system/stats")
	if err != nil {
		return 0, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("нет связи с панелью: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, classifyHTTP(resp)
	}
	var out struct {
		Response struct {
			Users struct {
				TotalUsers int `json:"totalUsers"`
			} `json:"users"`
		} `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, fmt.Errorf("разбор ответа панели: %w", err)
	}
	return out.Response.Users.TotalUsers, nil
}

// classifyHTTP превращает не-200 ответ в понятную пользователю ошибку,
// подсказывая вероятную причину (токен/кука/защита /api).
func classifyHTTP(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	snippet := strings.TrimSpace(string(body))
	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("панель отклонила доступ (HTTP %d): проверьте API-token. %s",
			resp.StatusCode, snippet)
	case http.StatusNotFound:
		return fmt.Errorf("эндпоинт не найден (HTTP 404): проверьте URL панели")
	default:
		return fmt.Errorf("панель вернула HTTP %d: %s", resp.StatusCode, snippet)
	}
}
