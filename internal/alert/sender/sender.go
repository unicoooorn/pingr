package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/unicoooorn/pingr/internal/service"
)

var _ service.AlertSender = &tgApi{}

type tgApi struct {
	url    string
	token  string
	chatId string
}

func NewTgApi(url string, token string, chatId string) *tgApi {
	return &tgApi{
		url:    url,
		token:  token,
		chatId: chatId,
	}
}

type sendPayload struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
}

func (l *tgApi) SendAlert(
	ctx context.Context,
	alertMessage string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	base := strings.TrimRight(l.url, "/")
	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", base, l.token)

	payload := sendPayload{
		ChatID:                l.chatId,
		Text:                  alertMessage,
		ParseMode:             "",
		DisableWebPagePreview: true,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal tg payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("create request to tg api: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do request to tg api: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tg api returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}
