package aws

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

// CostClient はAWS Cost Explorer APIのクライアント
type CostClient struct {
	client *costexplorer.Client
}

// NewCostClient は新しいCostClientを作成します
func NewCostClient(ctx context.Context, region, accessKeyID, secretAccessKey string) (*CostClient, error) {
	var opts []func(*config.LoadOptions) error

	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	// アクセスキーとシークレットキーが指定されている場合は静的認証情報を使用
	if accessKeyID != "" && secretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &CostClient{
		client: costexplorer.NewFromConfig(cfg),
	}, nil
}

// DailyCostSummary は1日分のコストサマリー
type DailyCostSummary struct {
	Date       string
	TotalCost  float64
	Currency   string
	Services   []ServiceCost
}

// ServiceCost はサービス別のコスト
type ServiceCost struct {
	Service string
	Cost    float64
}

// MonthlyCostSummary は月間コストサマリー
type MonthlyCostSummary struct {
	MonthToDate float64
	Forecast    float64
	Currency    string
}

// GetYesterdayCost は前日のコストを取得します
func (c *CostClient) GetYesterdayCost(ctx context.Context) (*DailyCostSummary, error) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	return c.getDailyCost(ctx, yesterday)
}

// getDailyCost は指定日のコストを取得します
func (c *CostClient) getDailyCost(ctx context.Context, date time.Time) (*DailyCostSummary, error) {
	start := date.Format("2006-01-02")
	end := date.AddDate(0, 0, 1).Format("2006-01-02")

	// 合計コストを取得
	totalInput := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: types.GranularityDaily,
		Metrics: []string{
			"UnblendedCost",
		},
	}

	totalResult, err := c.client.GetCostAndUsage(ctx, totalInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get total cost: %w", err)
	}

	var totalCost float64
	var currency string

	if len(totalResult.ResultsByTime) > 0 && len(totalResult.ResultsByTime[0].Total) > 0 {
		if cost, ok := totalResult.ResultsByTime[0].Total["UnblendedCost"]; ok {
			if cost.Amount != nil {
				fmt.Sscanf(*cost.Amount, "%f", &totalCost)
			}
			if cost.Unit != nil {
				currency = *cost.Unit
			}
		}
	}

	// サービス別コストを取得
	serviceInput := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: types.GranularityDaily,
		Metrics: []string{
			"UnblendedCost",
		},
		GroupBy: []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
		},
	}

	serviceResult, err := c.client.GetCostAndUsage(ctx, serviceInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get service costs: %w", err)
	}

	var services []ServiceCost

	if len(serviceResult.ResultsByTime) > 0 {
		for _, group := range serviceResult.ResultsByTime[0].Groups {
			if len(group.Keys) > 0 && len(group.Metrics) > 0 {
				serviceName := group.Keys[0]
				if cost, ok := group.Metrics["UnblendedCost"]; ok {
					if cost.Amount != nil {
						var amount float64
						fmt.Sscanf(*cost.Amount, "%f", &amount)
						if amount > 0 {
							services = append(services, ServiceCost{
								Service: serviceName,
								Cost:    amount,
							})
						}
					}
				}
			}
		}
	}

	// コストの高い順にソート
	sort.Slice(services, func(i, j int) bool {
		return services[i].Cost > services[j].Cost
	})

	return &DailyCostSummary{
		Date:      start,
		TotalCost: totalCost,
		Currency:  currency,
		Services:  services,
	}, nil
}

// GetMonthlyForecast は当月の予測コストを取得します
func (c *CostClient) GetMonthlyForecast(ctx context.Context) (*MonthlyCostSummary, error) {
	now := time.Now()

	// 月初から今日までの実績コスト
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	today := now.Format("2006-01-02")

	mtdInput := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(monthStart.Format("2006-01-02")),
			End:   aws.String(today),
		},
		Granularity: types.GranularityMonthly,
		Metrics: []string{
			"UnblendedCost",
		},
	}

	mtdResult, err := c.client.GetCostAndUsage(ctx, mtdInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get month-to-date cost: %w", err)
	}

	var monthToDate float64
	var currency string

	if len(mtdResult.ResultsByTime) > 0 && len(mtdResult.ResultsByTime[0].Total) > 0 {
		if cost, ok := mtdResult.ResultsByTime[0].Total["UnblendedCost"]; ok {
			if cost.Amount != nil {
				fmt.Sscanf(*cost.Amount, "%f", &monthToDate)
			}
			if cost.Unit != nil {
				currency = *cost.Unit
			}
		}
	}

	// 月末までの予測コスト
	monthEnd := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())

	forecastInput := &costexplorer.GetCostForecastInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(today),
			End:   aws.String(monthEnd.Format("2006-01-02")),
		},
		Granularity: types.GranularityMonthly,
		Metric:      types.MetricUnblendedCost,
	}

	forecastResult, err := c.client.GetCostForecast(ctx, forecastInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost forecast: %w", err)
	}

	var forecastAmount float64
	if forecastResult.Total != nil && forecastResult.Total.Amount != nil {
		fmt.Sscanf(*forecastResult.Total.Amount, "%f", &forecastAmount)
	}

	return &MonthlyCostSummary{
		MonthToDate: monthToDate,
		Forecast:    monthToDate + forecastAmount,
		Currency:    currency,
	}, nil
}
