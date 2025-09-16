package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"lumina/internal/agent"
)

var serveCommand = &cobra.Command{
	Use:   "serve",
	Short: "Start lumina agent",
	Run: func(cmd *cobra.Command, args []string) {
		runServe()
	},
}

func init() {
}

func runServe() {
	conf, err := agent.LoadConfig(configFile)
	if err != nil {
		logrus.Fatal("initConfig error, ", err.Error())
	}

	logrus.Infof("config: %+v", conf)

	agent, err := agent.NewAgent(conf)
	if err != nil {
		logrus.WithError(err).Fatalf("new agent")
		return
	}
	go agent.Start()

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	<-termChan
	logrus.Infof("agent is shutting down...")
	agent.Stop()
}
