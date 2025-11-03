# AWS Cost Notifier

AWS Cost Explorer APIを使用してAWSコストを取得し、Slackに通知するGoアプリケーションです。
日次でCronから実行することで、前日のコスト、サービス別使用状況、当月の累計コストを自動通知できます。

[English README is here](README.en.md)

## ⚠️ 重要: API利用料金について

**このツールを1回実行するたびに約$0.02のAWS料金が発生します。**

AWS Cost Explorer APIには以下の料金がかかります：
- `GetCostAndUsage` API: **$0.01/リクエスト**
- このツールは1回の実行で2回APIを呼び出します（前日のコスト + 月間累計）

**月間コスト試算**:
- 毎日1回実行: 約 **$0.60/月**（30日）
- 毎日1回実行: 約 **$7.30/年**（365日）

コストを抑えたい場合は、実行頻度を週1回にする、または閾値を上げて通知をスキップする頻度を増やすなどの調整をご検討ください。

詳細: [AWS Cost Explorer 料金](https://aws.amazon.com/jp/aws-cost-management/aws-cost-explorer/pricing/)

## 主な機能

- 前日のAWSコストを取得
- サービス別のコスト内訳を表示（上位10件）
- 当月の累計コストを取得
- Slack Webhook経由でリッチな通知を送信
- コスト金額に応じた色分け表示（緑/黄/赤）
- $0.01未満の微小コストは通知をスキップ
- Cron対応の設計で定期実行が可能

## 前提条件

- Docker Desktop または Docker Engine が開発用端末にインストールされていること
- VS Code が開発用端末にインストールされていること
- VS Code 拡張機能 **Dev Containers** がインストールされていること
- AWSアカウントとCost Explorer APIへのアクセス権限
- Slack Incoming Webhook URL

## セットアップ

### 1. リポジトリのクローン

```bash
git clone https://github.com/foresuke/aws-cost-notifier.git
cd aws-cost-notifier
```

### 2. Dev Containerで開く

1. VS Codeでプロジェクトフォルダを開く
2. コマンドパレット (`Ctrl+Shift+P` または `Cmd+Shift+P`) を開いて「Dev Containers: Reopen in Container」を実行
3. コンテナのビルドと起動を待つ

### 3. AWS認証情報の設定

AWS CLIの認証情報を設定します。以下のいずれかの方法を使用してください：

#### 方法A: AWS CLIで設定

```bash
aws configure
```

#### 方法B: 環境変数で設定

```bash
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_DEFAULT_REGION=us-east-1
```

#### 方法C: AWS CLIプロファイルを使用

```bash
# ~/.aws/credentials にプロファイルを追加
[my-profile]
aws_access_key_id = your_access_key
aws_secret_access_key = your_secret_key
```

### 4. 設定ファイルの編集

[config.toml](config.toml) を編集してSlack Webhook URLとAWS設定を記述します：

```toml
# アプリケーション名
app_name = "AWSコスト通知"

# 作者
author = "your_name"

# デバッグモードかどうか
debug = false

# AWS設定
[aws]
region = "us-east-1"           # AWSリージョン
access_key_id = ""             # AWSアクセスキーID
secret_access_key = ""         # AWSシークレットアクセスキー

# Slack通知設定
[slack]
enabled = true
webhook_url = "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
timeout = "10s"
```

### 5. ビルド

```bash
go build -o aws-cost-notifier .
```

## 使い方

### コマンド一覧

```bash
# コスト通知を実行
./aws-cost-notifier cost

# バージョン確認
./aws-cost-notifier version

# 設定確認
./aws-cost-notifier info

# Slack通知のテスト
./aws-cost-notifier notify --type info --message "テストメッセージ"
```

### コスト通知の実行

```bash
./aws-cost-notifier cost
```

このコマンドは以下の情報をSlackに通知します：

1. **前日のコスト**
   - 合計金額
   - サービス別の内訳（上位10件、コストの高い順）

2. **当月の累計**
   - 月初から現在までの累計コスト

通知メッセージはコスト金額に応じて色分けされます：
- 緑：$50未満
- 黄：$50以上$100未満
- 赤：$100以上

## Cronでの定期実行

### 日次実行の設定例

毎朝9時にコスト通知を実行する場合：

```bash
# crontabを編集
crontab -e

# 以下を追加（毎朝9:00に実行）
0 9 * * * /path/to/aws-cost-notifier cost >> /var/log/aws-cost-notifier.log 2>&1
```

### 環境変数を使用する場合

Cronでは環境変数が限定的なため、スクリプトを作成することを推奨します：

```bash
#!/bin/bash
# /path/to/run-cost-notifier.sh

# AWS認証情報
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_DEFAULT_REGION=us-east-1

# または AWS CLIプロファイルを使用
export AWS_PROFILE=my-profile

# 設定ファイルのパスを指定
cd /path/to/aws-cost-notifier
./aws-cost-notifier cost
```

```bash
# スクリプトに実行権限を付与
chmod +x /path/to/run-cost-notifier.sh

# crontab設定
0 9 * * * /path/to/run-cost-notifier.sh >> /var/log/aws-cost-notifier.log 2>&1
```

### systemd timerを使用する場合

より柔軟な設定が可能なsystemd timerを使用することもできます：

```ini
# /etc/systemd/system/aws-cost-notifier.service
[Unit]
Description=AWS Cost Notifier
After=network.target

[Service]
Type=oneshot
User=your_user
WorkingDirectory=/path/to/aws-cost-notifier
Environment="AWS_REGION=us-east-1"
Environment="AWS_PROFILE=my-profile"
ExecStart=/path/to/aws-cost-notifier cost
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

```ini
# /etc/systemd/system/aws-cost-notifier.timer
[Unit]
Description=AWS Cost Notifier Timer
Requires=aws-cost-notifier.service

[Timer]
OnCalendar=*-*-* 09:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

```bash
# timerを有効化して起動
sudo systemctl enable aws-cost-notifier.timer
sudo systemctl start aws-cost-notifier.timer

# 状態確認
sudo systemctl status aws-cost-notifier.timer
```

## 必要なAWS権限

Cost Explorer APIへのアクセスには以下のIAMポリシーが必要です：

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ce:GetCostAndUsage",
        "ce:GetCostForecast"
      ],
      "Resource": "*"
    }
  ]
}
```

## プロジェクト構成

```
aws-cost-notifier/
├── .devcontainer/          # Dev Containers設定
├── .github/workflows/      # GitHub Actions（自動ビルド・リリース）
├── cmd/                    # Cobraコマンド定義
│   ├── root.go            # ルートコマンド
│   ├── cost.go            # コスト通知コマンド
│   ├── info.go            # 設定表示コマンド
│   ├── version.go         # バージョン表示コマンド
│   └── notify.go          # Slack通知テストコマンド
├── pkg/
│   ├── aws/
│   │   └── cost.go        # AWS Cost Explorer APIクライアント
│   └── notifier/
│       ├── notifier.go    # 通知インターフェース
│       └── slack.go       # Slack通知実装
├── config.toml            # 設定ファイル
├── main.go                # エントリーポイント
├── go.mod                 # Go modules
└── README.md              # このファイル
```

## トラブルシューティング

### Cost Explorerのデータが取得できない

- AWSアカウントでCost Explorerが有効化されているか確認してください
- Cost Explorerのデータは24時間遅延するため、当日のデータは取得できません
- IAM権限が正しく設定されているか確認してください

### Slack通知が届かない

- Webhook URLが正しいか確認してください
- [config.toml](config.toml) で `slack.enabled = true` になっているか確認してください
- `notify` コマンドでテスト通知を送信してみてください：
  ```bash
  ./aws-cost-notifier notify --type info --message "テスト"
  ```

### 認証エラーが発生する

- AWS認証情報が正しく設定されているか確認してください
- 環境変数 `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` が設定されているか確認してください
- [config.toml](config.toml) の `aws.access_key_id` と `aws.secret_access_key` が正しく設定されているか確認してください

## 開発

### テスト実行

```bash
go test ./...
```

### リリース

git タグをプッシュすると、GitHub Actionsが自動的に各プラットフォーム向けのバイナリをビルドしてリリースします：

```bash
git tag v1.0.0
git push origin v1.0.0
```

## ライセンス

MIT License
