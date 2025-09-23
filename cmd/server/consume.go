package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"lumina/internal/consumer"
	"lumina/internal/model"
)

var consumeCmd = &cobra.Command{
	Use:   "consume",
	Short: "Consume messages from NSQ",
	Long:  `Consume messages from NSQ`,
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := consumer.LoadConfig(configFile)
		if err != nil {
			logrus.Fatal("initConfig error, ", err.Error())
		}

		db, err := model.InitDB(conf.DB)
		if err != nil {
			logrus.Fatal("failed to init database", err)
		}
		defer func() {
			sqlDB, _ := db.DB()
			sqlDB.Close()
		}()

		c, err := consumer.NewConsumer(conf)
		if err != nil {
			logrus.Fatalf("Failed to create consumer: %v", err)
		}
		go c.Start()

		termChan := make(chan os.Signal, 1)
		signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

		<-termChan
		logrus.Infof("consumer is shutting down...")
		c.Stop()
	},
}

func init() {
}
