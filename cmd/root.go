package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"lumina/internal/version"
	"lumina/pkg/log"
)

var (
	logLevel   string
	configFile string
)

var rootCmd = &cobra.Command{
	Use:   "lumina",
	Short: "lumina is a AI vision engine",
	Long: `A Fast and Easy-to-Use AI Vision Engine.
Version: ` + version.VERSION + `/` + version.COMMIT,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.InitLog(logLevel)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "etc/config.yaml", "Path to config file")

	rootCmd.AddCommand(serveCommand)
	rootCmd.AddCommand(updateDBCommand)
	rootCmd.AddCommand(toolsCmd)
}
