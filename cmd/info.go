package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// sensitiveKeywords は機密情報を含む可能性のある設定キーのキーワードリストです。
// これらのキーワード（大文字・小文字を区別しない）がキーに含まれている場合、その値はマスクされます。
var sensitiveKeywords = []string{
	"password",
	"passphrase",
	"secret",
	"token",
	"apikey",
	"access_key",
	"secret_key",
	"private_key",
	"credential",
}

const maskText = "****"

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display application configuration",
	Long:  `This command outputs the currently loaded configuration settings in JSON format. Sensitive information will be masked.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Viperから全ての設定を取得
		allSettings := viper.AllSettings()
		// 設定情報を再帰的にマスク処理
		maskedSettings := maskSensitiveDataRecursive(allSettings)

		// マスク処理された設定をJSON形式に変換
		jsonData, err := json.MarshalIndent(maskedSettings, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshalling settings to JSON: %s\n", err)
			os.Exit(1)
		}

		// JSONデータを標準出力
		fmt.Println(string(jsonData))
	},
}

func init() {
}

// maskSensitiveDataRecursive は、設定情報を再帰的に処理し、機密情報を含むキーの値をマスクします。
func maskSensitiveDataRecursive(data map[string]interface{}) map[string]interface{} {
	maskedData := make(map[string]interface{})
	for key, value := range data {
		isSensitiveKey := false
		lowerKey := strings.ToLower(key)
		for _, keyword := range sensitiveKeywords {
			if strings.Contains(lowerKey, keyword) {
				isSensitiveKey = true
				break
			}
		}

		if isSensitiveKey {
			maskedData[key] = maskText
		} else {
			// 値がマップの場合、再帰的に処理
			if subMap, ok := value.(map[string]interface{}); ok {
				maskedData[key] = maskSensitiveDataRecursive(subMap)
			} else {
				maskedData[key] = value
			}
		}
	}
	return maskedData
}
