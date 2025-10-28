package notifier

import "context"

// Notifier は通知を送信するための共通インターフェース
type Notifier interface {
	// 基本的なメッセージ送信
	SendMessage(ctx context.Context, msg string, opts ...Option) error

	// レベル別の送信メソッド
	SendInfo(ctx context.Context, msg string, opts ...Option) error
	SendWarning(ctx context.Context, msg string, opts ...Option) error
	SendError(ctx context.Context, msg string, opts ...Option) error
	SendSuccess(ctx context.Context, msg string, opts ...Option) error

	// リッチメッセージ
	SendAttachment(ctx context.Context, att *Attachment) error

	// 複数のAttachmentを含むメッセージを送信
	Send(ctx context.Context, msg Message) error

	// 有効/無効チェック
	IsEnabled() bool
}

// Attachment はリッチな通知のための構造体
type Attachment struct {
	Title     string
	Text      string
	Color     string // good, warning, danger, または16進数カラー
	Fields    []Field
	Footer    string
	Timestamp int64
}

// Field は添付ファイルのフィールド
type Field struct {
	Title string
	Value string
	Short bool // 短い表示（2カラム）か
}

// AttachmentField は Field のエイリアス
type AttachmentField = Field

// Message は複数のAttachmentを含むメッセージ
type Message struct {
	Text        string
	Attachments []Attachment
}

// Option は送信オプションを設定する関数型
type Option func(*MessageOptions)

// MessageOptions は送信オプション
type MessageOptions struct {
	Channel   string
	Username  string
	IconEmoji string
	IconURL   string
	Mentions  []string // メンション対象（@user, @channel等）
}

// WithChannel はチャンネルを指定するオプション
func WithChannel(channel string) Option {
	return func(o *MessageOptions) {
		o.Channel = channel
	}
}

// WithUsername はユーザー名を指定するオプション
func WithUsername(username string) Option {
	return func(o *MessageOptions) {
		o.Username = username
	}
}

// WithIconEmoji はアイコン絵文字を指定するオプション
func WithIconEmoji(emoji string) Option {
	return func(o *MessageOptions) {
		o.IconEmoji = emoji
	}
}

// WithIconURL はアイコンURLを指定するオプション
func WithIconURL(url string) Option {
	return func(o *MessageOptions) {
		o.IconURL = url
	}
}

// WithMentions はメンションを指定するオプション
func WithMentions(mentions ...string) Option {
	return func(o *MessageOptions) {
		o.Mentions = mentions
	}
}

// NullNotifier は何もしない通知器（無効時に使用）
type NullNotifier struct{}

// NewNull は新しいNullNotifierを作成
func NewNull() *NullNotifier {
	return &NullNotifier{}
}

// NewNullNotifier は新しいNullNotifierを作成（エイリアス）
func NewNullNotifier() *NullNotifier {
	return &NullNotifier{}
}

func (n *NullNotifier) SendMessage(ctx context.Context, msg string, opts ...Option) error {
	return nil
}
func (n *NullNotifier) SendInfo(ctx context.Context, msg string, opts ...Option) error { return nil }
func (n *NullNotifier) SendWarning(ctx context.Context, msg string, opts ...Option) error {
	return nil
}
func (n *NullNotifier) SendError(ctx context.Context, msg string, opts ...Option) error { return nil }
func (n *NullNotifier) SendSuccess(ctx context.Context, msg string, opts ...Option) error {
	return nil
}
func (n *NullNotifier) SendAttachment(ctx context.Context, att *Attachment) error { return nil }
func (n *NullNotifier) Send(ctx context.Context, msg Message) error               { return nil }
func (n *NullNotifier) IsEnabled() bool                                           { return false }
