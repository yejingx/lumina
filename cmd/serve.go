package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"lumina/internal/config"
	"lumina/internal/model"
	"lumina/internal/server"
)

var serveCommand = &cobra.Command{
	Use:   "serve",
	Short: "Start lumina server",
	Run: func(cmd *cobra.Command, args []string) {
		runServe()
	},
}

func init() {
}

func runServe() {
	conf, err := config.InitConfig(configFile)
	if err != nil {
		logrus.Fatal("initConfig error, ", err.Error())
	}

	logrus.Infof("config: %+v", conf)

	db, err := model.InitDB(conf.DB)
	if err != nil {
		logrus.Fatal("failed to init database", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	ctx, cancelFunc := context.WithCancel(context.Background())

	srv, err := server.NewServer(ctx, conf)
	if err != nil {
		logrus.Fatalf("newServer error, %s", err.Error())
		cancelFunc()
		return
	}
	go srv.Start()

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	<-termChan
	logrus.Infof("server is shutting down...")
	srv.Shutdown()
	cancelFunc()
}
