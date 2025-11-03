# AWS Cost Notifier

A Go application that retrieves AWS costs using the Cost Explorer API and sends notifications to Slack.
When run daily via Cron, it automatically notifies you of the previous day's costs, service-by-service breakdown, and current month's cumulative costs.

[日本語版 README はこちら](README.md)

## ⚠️ Important: API Usage Costs

**Each execution of this tool incurs approximately $0.02 in AWS charges.**

AWS Cost Explorer API pricing:
- `GetCostAndUsage` API: **$0.01/request**
- This tool makes 2 API calls per execution (previous day's cost + monthly cumulative)

**Monthly cost estimates**:
- Daily execution: Approximately **$0.60/month** (30 days)
- Daily execution: Approximately **$7.30/year** (365 days)

To reduce costs, consider adjusting the execution frequency to weekly, or raising the threshold to skip notifications more often.

Details: [AWS Cost Explorer Pricing](https://aws.amazon.com/aws-cost-management/aws-cost-explorer/pricing/)

## Features

- Retrieve previous day's AWS costs
- Display service-by-service cost breakdown (top 10)
- Retrieve current month's cumulative costs
- Send rich notifications via Slack Webhook
- Color-coded display based on cost amount (green/yellow/red)
- Skip notifications for costs under $0.01
- Designed for scheduled execution via Cron

## Prerequisites

- Docker Desktop or Docker Engine installed on your development machine
- VS Code installed on your development machine
- VS Code **Dev Containers** extension installed
- AWS account with Cost Explorer API access permissions
- Slack Incoming Webhook URL

## Setup

### 1. Clone the Repository

```bash
git clone https://github.com/foresuke/aws-cost-notifier.git
cd aws-cost-notifier
```

### 2. Open in Dev Container

1. Open the project folder in VS Code
2. Open Command Palette (`Ctrl+Shift+P` or `Cmd+Shift+P`) and run "Dev Containers: Reopen in Container"
3. Wait for the container to build and start

### 3. Configure AWS Credentials

Set up AWS CLI credentials using one of the following methods:

#### Method A: Using AWS CLI

```bash
aws configure
```

#### Method B: Using Environment Variables

```bash
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_DEFAULT_REGION=us-east-1
```

#### Method C: Using AWS CLI Profile

```bash
# Add profile to ~/.aws/credentials
[my-profile]
aws_access_key_id = your_access_key
aws_secret_access_key = your_secret_key
```

### 4. Edit Configuration File

Edit [config.toml](config.toml) to add your Slack Webhook URL and AWS settings:

```toml
# Application name
app_name = "AWS Cost Notifier"

# Author
author = "your_name"

# Debug mode
debug = false

# AWS settings
[aws]
region = "us-east-1"           # AWS region
access_key_id = ""             # AWS access key ID
secret_access_key = ""         # AWS secret access key

# Slack notification settings
[slack]
enabled = true
webhook_url = "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
timeout = "10s"
```

### 5. Build

```bash
go build -o aws-cost-notifier .
```

## Usage

### Available Commands

```bash
# Execute cost notification
./aws-cost-notifier cost

# Check version
./aws-cost-notifier version

# Check configuration
./aws-cost-notifier info

# Test Slack notification
./aws-cost-notifier notify --type info --message "Test message"
```

### Execute Cost Notification

```bash
./aws-cost-notifier cost
```

This command sends the following information to Slack:

1. **Previous Day's Cost**
   - Total amount
   - Service-by-service breakdown (top 10, sorted by cost)

2. **Current Month's Cumulative**
   - Cumulative cost from beginning of month to present

Notification messages are color-coded based on cost amount:
- Green: Under $50
- Yellow: $50 or more, under $100
- Red: $100 or more

## Scheduled Execution with Cron

### Daily Execution Example

To execute cost notification at 9 AM every day:

```bash
# Edit crontab
crontab -e

# Add the following (executes at 9:00 AM daily)
0 9 * * * /path/to/aws-cost-notifier cost >> /var/log/aws-cost-notifier.log 2>&1
```

### Using Environment Variables

Since Cron has limited environment variables, we recommend creating a script:

```bash
#!/bin/bash
# /path/to/run-cost-notifier.sh

# AWS credentials
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_DEFAULT_REGION=us-east-1

# Or use AWS CLI profile
export AWS_PROFILE=my-profile

# Specify config file path
cd /path/to/aws-cost-notifier
./aws-cost-notifier cost
```

```bash
# Grant execute permission to script
chmod +x /path/to/run-cost-notifier.sh

# Crontab configuration
0 9 * * * /path/to/run-cost-notifier.sh >> /var/log/aws-cost-notifier.log 2>&1
```

### Using systemd timer

You can also use systemd timer for more flexible configuration:

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
# Enable and start timer
sudo systemctl enable aws-cost-notifier.timer
sudo systemctl start aws-cost-notifier.timer

# Check status
sudo systemctl status aws-cost-notifier.timer
```

## Required AWS Permissions

The following IAM policy is required for Cost Explorer API access:

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

## Project Structure

```
aws-cost-notifier/
├── .devcontainer/          # Dev Containers configuration
├── .github/workflows/      # GitHub Actions (automated build & release)
├── cmd/                    # Cobra command definitions
│   ├── root.go            # Root command
│   ├── cost.go            # Cost notification command
│   ├── info.go            # Configuration display command
│   ├── version.go         # Version display command
│   └── notify.go          # Slack notification test command
├── pkg/
│   ├── aws/
│   │   └── cost.go        # AWS Cost Explorer API client
│   └── notifier/
│       ├── notifier.go    # Notifier interface
│       └── slack.go       # Slack notification implementation
├── config.toml            # Configuration file
├── main.go                # Entry point
├── go.mod                 # Go modules
└── README.md              # This file
```

## Troubleshooting

### Cannot Retrieve Cost Explorer Data

- Verify that Cost Explorer is enabled in your AWS account
- Cost Explorer data has a 24-hour delay, so current day's data cannot be retrieved
- Verify that IAM permissions are correctly configured

### Slack Notifications Not Arriving

- Verify that the Webhook URL is correct
- Check that `slack.enabled = true` in [config.toml](config.toml)
- Try sending a test notification with the `notify` command:
  ```bash
  ./aws-cost-notifier notify --type info --message "Test"
  ```

### Authentication Errors Occurring

- Verify that AWS credentials are correctly configured
- Check that environment variables `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` are set
- If using AWS CLI profile, configure `aws.profile` in [config.toml](config.toml)

## Development

### Run Tests

```bash
go test ./...
```

### Release

When you push a git tag, GitHub Actions automatically builds binaries for each platform and creates a release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## License

MIT License
