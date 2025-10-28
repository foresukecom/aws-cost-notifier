package cmd

import (
	"fmt"
	"runtime" // Goのバージョンも表示するために追加

	"github.com/spf13/cobra"
)

// Version はアプリケーションのバージョンを保持します。
// ビルド時に goreleaser の ldflags によって設定されます。
// ★★★ 変数宣言のみにする ★★★
var Version string

// CommitHash はビルド時のコミットハッシュを保持します。
// ビルド時に goreleaser の ldflags によって設定されます。
// ★★★ コミットハッシュ用の変数を追加 ★★★
var CommitHash string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the application version",
	Long:  `This command prints the version number and build information of the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// アプリケーション名とバージョンを出力
		fmt.Printf("%s version: %s\n", rootCmd.Use, Version)

		// ★★★ コミットハッシュとGoのバージョンも出力 ★★★
		if CommitHash != "" {
			fmt.Printf("Commit: %s\n", CommitHash)
		}
		fmt.Printf("Built with Go: %s\n", runtime.Version())
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
