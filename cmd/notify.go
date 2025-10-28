package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/foresuke/aws-cost-notifier/pkg/notifier"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Slack通知機能のデモコマンド",
	Long: `Slack通知機能をテストするためのデモコマンドです。
各種通知タイプ（info/warning/error/success）を試すことができます。`,
	RunE: runNotify,
}

var (
	notifyType    string
	notifyMessage string
)

func init() {
	// フラグの追加
	notifyCmd.Flags().StringVarP(&notifyType, "type", "t", "info", "通知タイプ (info/warning/error/success/rich)")
	notifyCmd.Flags().StringVarP(&notifyMessage, "message", "m", "", "送信するメッセージ")

	// rootCmdに追加
	rootCmd.AddCommand(notifyCmd)
}

func runNotify(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Notifierの初期化
	n := initNotifier()

	if !n.IsEnabled() {
		fmt.Println("⚠️  Slack通知が無効です。config.tomlで設定を有効にしてください。")
		fmt.Println()
		fmt.Println("設定例:")
		fmt.Println("  [slack]")
		fmt.Println("  enabled = true")
		fmt.Println("  webhook_url = \"https://hooks.slack.com/services/YOUR/WEBHOOK/URL\"")
		return nil
	}

	// メッセージのデフォルト値
	if notifyMessage == "" {
		notifyMessage = fmt.Sprintf("テスト通知: %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("📤 Slack通知を送信中... (type: %s)\n", notifyType)

	var err error
	switch notifyType {
	case "info":
		err = n.SendInfo(ctx, notifyMessage)
	case "warning":
		err = n.SendWarning(ctx, notifyMessage)
	case "error":
		err = n.SendError(ctx, notifyMessage)
	case "success":
		err = n.SendSuccess(ctx, notifyMessage)
	case "rich":
		// リッチな通知例
		err = n.SendAttachment(ctx, &notifier.Attachment{
			Title: "リッチ通知のデモ",
			Text:  "これはリッチな通知のサンプルです",
			Color: "good",
			Fields: []notifier.Field{
				{Title: "実行時刻", Value: time.Now().Format("15:04:05"), Short: true},
				{Title: "実行ユーザー", Value: "CLI Bot", Short: true},
				{Title: "メッセージ", Value: notifyMessage, Short: false},
			},
			Footer:    "Go CLI App",
			Timestamp: time.Now().Unix(),
		})
	default:
		return fmt.Errorf("不明な通知タイプ: %s (使用可能: info/warning/error/success/rich)", notifyType)
	}

	if err != nil {
		fmt.Printf("❌ 通知の送信に失敗しました: %v\n", err)
		return err
	}

	fmt.Println("✅ 通知を送信しました")
	return nil
}

// initNotifier はViperの設定からNotifierを初期化
func initNotifier() notifier.Notifier {
	// Slack通知が無効の場合はNullNotifierを返す
	if !viper.GetBool("slack.enabled") {
		return notifier.NewNull()
	}

	// タイムアウトの取得
	timeout := 10 * time.Second
	if timeoutStr := viper.GetString("slack.timeout"); timeoutStr != "" {
		if d, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = d
		}
	}

	// SlackNotifierの作成
	return notifier.NewSlack(notifier.SlackConfig{
		WebhookURL: viper.GetString("slack.webhook_url"),
		AppName:    viper.GetString("app_name"),
		Enabled:    viper.GetBool("slack.enabled"),
		Timeout:    timeout,
	})
}
