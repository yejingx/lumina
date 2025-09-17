package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"lumina/internal/agent"
)

var registerCmd = &cobra.Command{
	Use:   "register <path to agent-info.json>",
	Short: "Register the agent",
	Long:  `Register the agent from json file`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		registerAgent(args[0])
	},
}

func registerAgent(agentInfoPath string) {
	logrus.Infof("Registering agent from file: %s", agentInfoPath)
	conf, err := agent.LoadConfig(configFile)
	if err != nil {
		logrus.Fatal("initConfig error, ", err.Error())
	}

	data, err := os.ReadFile(agentInfoPath)
	if err != nil {
		logrus.WithError(err).Fatalf("read agent info file")
		return
	}

	var agentInfo agent.AgentInfo
	if err2 := json.Unmarshal(data, &agentInfo); err2 != nil {
		logrus.WithError(err2).Fatalf("unmarshal agent info from file")
		return
	}
	registerTime := time.Now().Format(time.RFC3339)
	agentInfo.RegisterTime = &registerTime

	metadataDB, err := agent.NewMetadataDB(conf.DataDir(), logrus.WithField("component", "metadataDB"))
	if err != nil {
		logrus.WithError(err).Fatalf("new metadata db")
		return
	}
	defer metadataDB.Close()

	if err := metadataDB.UpdateAgentInfo(&agentInfo); err != nil {
		logrus.WithError(err).Fatalf("update agent info in metadata db")
		return
	}

	logrus.Infof("register agent %s success", *agentInfo.Uuid)
}
