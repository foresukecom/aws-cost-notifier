package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	awscost "github.com/foresuke/aws-cost-notifier/pkg/aws"
	"github.com/foresuke/aws-cost-notifier/pkg/notifier"
)

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "AWS コストを取得してSlackに通知します",
	Long:  `前日のコスト、サービス別の使用状況、当月の予測コストをSlackに通知します。`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// AWS設定を取得
		region := viper.GetString("aws.region")
		accessKeyID := viper.GetString("aws.access_key_id")
		secretAccessKey := viper.GetString("aws.secret_access_key")

		// Cost Explorerクライアントを作成
		costClient, err := awscost.NewCostClient(ctx, region, accessKeyID, secretAccessKey)
		if err != nil {
			log.Fatalf("AWSクライアントの初期化に失敗しました: %v", err)
		}

		// 前日のコストを取得
		dailyCost, err := costClient.GetYesterdayCost(ctx)
		if err != nil {
			log.Fatalf("前日のコスト取得に失敗しました: %v", err)
		}

		// デバッグ: 実際のコスト金額を出力
		if viper.GetBool("debug") {
			fmt.Printf("[DEBUG] 前日のコスト: $%.6f %s\n", dailyCost.TotalCost, dailyCost.Currency)
		}

		// 前日のコストが0.01ドル未満の場合は通知をスキップ
		// (AWS Cost Explorerは非常に小さい金額を返すことがあるため)
		if dailyCost.TotalCost < 0.01 {
			fmt.Printf("前日のコストが$%.4fと小さいため、通知をスキップしました\n", dailyCost.TotalCost)
			return
		}

		// 月間予測を取得
		monthlyCost, err := costClient.GetMonthlyForecast(ctx)
		if err != nil {
			log.Fatalf("月間予測の取得に失敗しました: %v", err)
		}

		// Slack通知を作成
		slackNotifier := createNotifier()

		// 通知メッセージを送信
		if err := sendCostNotification(slackNotifier, dailyCost, monthlyCost); err != nil {
			log.Fatalf("通知の送信に失敗しました: %v", err)
		}

		fmt.Println("AWS コスト情報をSlackに通知しました")
	},
}

func init() {
	rootCmd.AddCommand(costCmd)
}

// createNotifier はSlack通知クライアントを作成します
func createNotifier() notifier.Notifier {
	enabled := viper.GetBool("slack.enabled")
	if !enabled {
		return notifier.NewNullNotifier()
	}

	webhookURL := viper.GetString("slack.webhook_url")
	timeoutStr := viper.GetString("slack.timeout")

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeout = 10 * time.Second
	}

	slack, err := notifier.NewSlackNotifier(webhookURL, timeout)
	if err != nil {
		log.Printf("Slack通知の初期化に失敗しました: %v", err)
		return notifier.NewNullNotifier()
	}

	return slack
}

// sendCostNotification はコスト情報をSlackに送信します
func sendCostNotification(n notifier.Notifier, daily *awscost.DailyCostSummary, monthly *awscost.MonthlyCostSummary) error {
	ctx := context.Background()

	// サービス別コストの詳細を構築（上位10件まで）
	var serviceFields []notifier.AttachmentField
	maxServices := 10
	if len(daily.Services) < maxServices {
		maxServices = len(daily.Services)
	}

	for i := 0; i < maxServices; i++ {
		service := daily.Services[i]
		serviceFields = append(serviceFields, notifier.AttachmentField{
			Title: service.Service,
			Value: fmt.Sprintf("$%.2f", service.Cost),
			Short: true,
		})
	}

	// 日次コストの通知
	dailyColor := getColorByAmount(daily.TotalCost)
	dailyAttachment := notifier.Attachment{
		Color:  dailyColor,
		Title:  fmt.Sprintf("前日のコスト (%s)", daily.Date),
		Text:   fmt.Sprintf("*$%.2f* %s", daily.TotalCost, daily.Currency),
		Fields: serviceFields,
	}

	// 月間累計の通知
	monthlyColor := getColorByAmount(monthly.MonthToDate)
	monthlyAttachment := notifier.Attachment{
		Color: monthlyColor,
		Title: "当月の累計",
		Fields: []notifier.AttachmentField{
			{
				Title: "月初から現在まで",
				Value: fmt.Sprintf("$%.2f", monthly.MonthToDate),
				Short: false,
			},
		},
	}

	message := notifier.Message{
		Text: "AWS コストレポート",
		Attachments: []notifier.Attachment{
			dailyAttachment,
			monthlyAttachment,
		},
	}

	return n.Send(ctx, message)
}

// getColorByAmount はコスト金額に応じて色を返します
func getColorByAmount(amount float64) string {
	switch {
	case amount >= 100:
		return "danger" // 赤
	case amount >= 50:
		return "warning" // 黄色
	default:
		return "good" // 緑
	}
}
