package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackNotifier はSlackへの通知を行う
type SlackNotifier struct {
	webhookURL string
	appName    string
	enabled    bool
	httpClient *http.Client
}

// SlackConfig はSlack通知の設定
type SlackConfig struct {
	WebhookURL string
	AppName    string
	Enabled    bool
	Timeout    time.Duration
}

// NewSlack は新しいSlackNotifierを作成
func NewSlack(config SlackConfig) *SlackNotifier {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	return &SlackNotifier{
		webhookURL: config.WebhookURL,
		appName:    config.AppName,
		enabled:    config.Enabled && config.WebhookURL != "",
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// NewSlackNotifier は新しいSlackNotifierを作成（簡易版）
func NewSlackNotifier(webhookURL string, timeout time.Duration) (*SlackNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}

	if timeout == 0 {
		timeout = 10 * time.Second
	}

	return &SlackNotifier{
		webhookURL: webhookURL,
		enabled:    true,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// IsEnabled は通知が有効かどうかを返す
func (s *SlackNotifier) IsEnabled() bool {
	return s.enabled
}

// SendMessage は基本的なメッセージを送信
func (s *SlackNotifier) SendMessage(ctx context.Context, msg string, opts ...Option) error {
	if !s.enabled {
		return nil
	}

	options := s.buildOptions(opts...)

	payload := map[string]any{
		"text": s.formatMessage(msg),
	}

	s.applyOptions(payload, options)

	return s.send(ctx, payload)
}

// SendInfo は情報メッセージを送信
func (s *SlackNotifier) SendInfo(ctx context.Context, msg string, opts ...Option) error {
	return s.SendAttachment(ctx, &Attachment{
		Text:  s.formatMessage(msg),
		Color: "#36a64f", // 緑
	})
}

// SendWarning は警告メッセージを送信
func (s *SlackNotifier) SendWarning(ctx context.Context, msg string, opts ...Option) error {
	return s.SendAttachment(ctx, &Attachment{
		Text:  s.formatMessage(msg),
		Color: "warning", // オレンジ
	})
}

// SendError はエラーメッセージを送信
func (s *SlackNotifier) SendError(ctx context.Context, msg string, opts ...Option) error {
	return s.SendAttachment(ctx, &Attachment{
		Text:  s.formatMessage(msg),
		Color: "danger", // 赤
	})
}

// SendSuccess は成功メッセージを送信
func (s *SlackNotifier) SendSuccess(ctx context.Context, msg string, opts ...Option) error {
	return s.SendAttachment(ctx, &Attachment{
		Text:  s.formatMessage(msg),
		Color: "good", // 緑
	})
}

// formatMessage はメッセージにアプリ名を付加
func (s *SlackNotifier) formatMessage(msg string) string {
	if s.appName != "" {
		return fmt.Sprintf("[%s] %s", s.appName, msg)
	}
	return msg
}

// SendAttachment はリッチな通知を送信
func (s *SlackNotifier) SendAttachment(ctx context.Context, att *Attachment) error {
	if !s.enabled {
		return nil
	}

	attachment := map[string]any{
		"fallback": att.Text,
		"text":     att.Text,
	}

	if att.Title != "" {
		attachment["title"] = att.Title
	}
	if att.Color != "" {
		attachment["color"] = att.Color
	}
	if att.Footer != "" {
		attachment["footer"] = att.Footer
	}
	if att.Timestamp > 0 {
		attachment["ts"] = att.Timestamp
	}
	if len(att.Fields) > 0 {
		fields := make([]map[string]any, len(att.Fields))
		for i, f := range att.Fields {
			fields[i] = map[string]any{
				"title": f.Title,
				"value": f.Value,
				"short": f.Short,
			}
		}
		attachment["fields"] = fields
	}

	payload := map[string]any{
		"attachments": []any{attachment},
	}

	return s.send(ctx, payload)
}

// Send は複数のAttachmentを含むメッセージを送信
func (s *SlackNotifier) Send(ctx context.Context, msg Message) error {
	if !s.enabled {
		return nil
	}

	payload := map[string]any{}

	if msg.Text != "" {
		payload["text"] = msg.Text
	}

	if len(msg.Attachments) > 0 {
		attachments := make([]map[string]any, len(msg.Attachments))
		for i, att := range msg.Attachments {
			attachment := map[string]any{
				"fallback": att.Text,
				"text":     att.Text,
			}

			if att.Title != "" {
				attachment["title"] = att.Title
			}
			if att.Color != "" {
				attachment["color"] = att.Color
			}
			if att.Footer != "" {
				attachment["footer"] = att.Footer
			}
			if att.Timestamp > 0 {
				attachment["ts"] = att.Timestamp
			}
			if len(att.Fields) > 0 {
				fields := make([]map[string]any, len(att.Fields))
				for j, f := range att.Fields {
					fields[j] = map[string]any{
						"title": f.Title,
						"value": f.Value,
						"short": f.Short,
					}
				}
				attachment["fields"] = fields
			}

			attachments[i] = attachment
		}
		payload["attachments"] = attachments
	}

	return s.send(ctx, payload)
}

// buildOptions はオプションを構築
func (s *SlackNotifier) buildOptions(opts ...Option) *MessageOptions {
	options := &MessageOptions{}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// applyOptions はpayloadにオプションを適用
func (s *SlackNotifier) applyOptions(payload map[string]any, opts *MessageOptions) {
	if opts.Username != "" {
		payload["username"] = opts.Username
	}
	if opts.IconEmoji != "" {
		payload["icon_emoji"] = opts.IconEmoji
	}
	if opts.IconURL != "" {
		payload["icon_url"] = opts.IconURL
	}

	// メンションを追加
	if len(opts.Mentions) > 0 {
		text := payload["text"].(string)
		for _, mention := range opts.Mentions {
			text = fmt.Sprintf("<%s> %s", mention, text)
		}
		payload["text"] = text
	}
}

// send は実際にSlackへリクエストを送信
func (s *SlackNotifier) send(ctx context.Context, payload map[string]any) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned status %d", resp.StatusCode)
	}

	return nil
}
