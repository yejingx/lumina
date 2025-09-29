package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"lumina/internal/device"
	"lumina/internal/device/config"
)

var serveCommand = &cobra.Command{
	Use:   "serve",
	Short: "Start lumina device",
	Run: func(cmd *cobra.Command, args []string) {
		runServe()
	},
}

func init() {
}

func runServe() {
	conf, err := config.LoadConfig(configFile)
	if err != nil {
		logrus.Fatal("initConfig error, ", err.Error())
	}

	logrus.Infof("config: %+v", conf)

	device, err := device.NewDevice(conf)
	if err != nil {
		logrus.WithError(err).Fatalf("new device")
		return
	}
	go device.Start()

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	<-termChan
	logrus.Infof("device is shutting down...")
	device.Stop()
}
