package main

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
	Use:   "lumina-device",
	Short: "lumina device is a vision AI task engine",
	Long: `lumina device is a vision AI device that runs on edge devices.
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
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "etc/device.yaml", "Path to config file")

	rootCmd.AddCommand(serveCommand)
	rootCmd.AddCommand(jobCmd)
	rootCmd.AddCommand(registerCmd)
}

func main() {
	Execute()
}
