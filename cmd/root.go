package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// cfgFile は設定ファイルのパスを保持するグローバル変数です
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "your_cli_app", // ★★★ ここをあなたのCLIアプリケーション名に置き換えてください ★★★
	Short: "A sample Go CLI application with Cobra and Viper",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

This application demonstrates basic CLI structure with Cobra and
configuration management using Viper.`, // 説明を少し変更
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// rootCmd にフラグや引数を追加する場所
	// persistent flags are global for the whole application
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is config.toml)")

	// Cobra が初期化される前に特定の関数を実行する設定
	cobra.OnInitialize(initViper)

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(infoCmd)
}

// initViper reads in config file and ENV variables if set.
func initViper() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in current directory and $HOME directory
		viper.AddConfigPath(".") // プロジェクトルート
		// viper.AddConfigPath("$HOME/.your_cli_app") // ★★★ 必要であればホームディレクトリ等も追加 ★★★
		viper.SetConfigName("config") // config.toml, config.json, config.yaml... を探す (拡張子なし)
		viper.SetConfigType("toml")   // TOML形式であることを明示的に指定
	}

	// 環境変数から設定を読み込む
	// 例: YOURAPP_GREETING_PREFIX="Hi, " で greeting.prefix が設定される
	viper.SetEnvPrefix("YOURAPP") // 環境変数名のプレフィックスを設定
	viper.AutomaticEnv()          // 環境変数から値を読み込む（プレフィックス付きまたは自動的にマッピング可能なもの）

	viper.SetDefault("debug", false) // デバッグモードのデフォルト

	// 設定ファイルを読み込む
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		// 設定ファイルが見つからない場合やその他の読み込みエラー
		// ConfigFileNotFoundError の場合は警告のみ、それ以外は Fatal
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore
			// fmt.Fprintln(os.Stderr, "Warning: Config file not found.")
		} else {
			// Config file was found but another error was produced
			fmt.Fprintf(os.Stderr, "Error reading config file: %s\n", err)
			// os.Exit(1) // 設定読み込みが必須ならここで終了
		}
	}
}
