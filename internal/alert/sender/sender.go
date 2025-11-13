package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"mime/multipart"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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
	infographics []byte,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Если картинки нет — старый путь: sendMessage JSON
	if len(infographics) == 0 {
		return l.sendMessage(ctx, alertMessage)
	}

	base := strings.TrimRight(l.url, "/")

	const telegramCaptionLimit = 1024
	caption := alertMessage
	truncated := false

	if len(caption) > telegramCaptionLimit {
		caption = caption[:telegramCaptionLimit]
		truncated = true
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendPhoto", base, l.token)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	// file part
	fw, err := w.CreateFormFile("photo", "infographic.png")
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := fw.Write(infographics); err != nil {
		return fmt.Errorf("write image to form: %w", err)
	}

	// fields
	if err := w.WriteField("chat_id", l.chatId); err != nil {
		return fmt.Errorf("write chat_id: %w", err)
	}
	if caption != "" {
		if err := w.WriteField("caption", caption); err != nil {
			return fmt.Errorf("write caption: %w", err)
		}
	}

	// обязательно закрыть writer перед созданием запроса — boundary финализируется
	if err := w.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return fmt.Errorf("create request to tg api: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do request to tg api: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tg api returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	// если подпись была обрезана — отправим полный текст отдельным sendMessage
	if truncated {
		if err := l.sendMessage(ctx, alertMessage); err != nil {
			return fmt.Errorf("sent photo but failed to send full message: %w", err)
		}
	}

	return nil
}

func (l *tgApi) sendMessage(ctx context.Context, text string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	base := strings.TrimRight(l.url, "/")
	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", base, l.token)

	payload := sendPayload{
		ChatID:                l.chatId,
		Text:                  text,
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

	client := &http.Client{Timeout: 15 * time.Second}
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

func (l *tgApi) Poll(ctx context.Context) (starts []string, stops []string, err error) {
	bot, err := tgbotapi.NewBotAPI(l.token)

	if err != nil {
		return nil, nil, fmt.Errorf("create bot api: %w", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := bot.GetUpdatesChan(u)
	defer bot.StopReceivingUpdates()

	return collectFromUpdates(ctx, updates)
}

func collectFromUpdates(ctx context.Context, updates <-chan tgbotapi.Update) (starts []string, stops []string, err error) {
	startSet := make(map[string]struct{})
	stopSet := make(map[string]struct{})

	for {
		select {
		case <-ctx.Done():
			return mapKeys(startSet), mapKeys(stopSet), ctx.Err()
		case upd, ok := <-updates:
			if !ok {
				return mapKeys(startSet), mapKeys(stopSet), errors.New("updates channel closed")
			}

			if upd.Message == nil {
				continue
			}

			chatID := upd.Message.Chat.ID
			chatIDStr := strconv.FormatInt(chatID, 10)

			text := strings.TrimSpace(upd.Message.Text)
			if text == "" {
				continue
			}
			if text[0] != '/' {
				continue
			}

			cmd := strings.ToLower(strings.TrimPrefix(strings.SplitN(text, " ", 2)[0], "/"))

			switch cmd {
			case "start":
				startSet[chatIDStr] = struct{}{}
			case "stop":
				stopSet[chatIDStr] = struct{}{}
			default:
				// ignore other commands
			}
		}
	}
}

func mapKeys(m map[string]struct{}) []string {
	res := make([]string, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	return res
}
