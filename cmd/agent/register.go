package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"lumina/internal/agent/config"
	"lumina/internal/agent/metadata"
	"lumina/internal/dao"
)

const (
	agentRegisterPath   = "/api/v1/agent/register"
	agentUnregisterPath = "/api/v1/agent/unregister"
)

var (
	serverAddr    = "http://localhost:8181"
	uuid          = ""
	showSensitive = false
	deviceName    = ""
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register tools",
	Long:  `Register tools`,
}

var showRegisterCmd = &cobra.Command{
	Use:   "show",
	Short: "Show registered agent info",
	Long:  `Show registered agent info`,
	Run: func(cmd *cobra.Command, args []string) {
		showRegisterInfo()
	},
}

var setRegisterCmd = &cobra.Command{
	Use:   "set",
	Short: "Set registered agent info",
	Long:  `Set registered agent info`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setRegisterInfo(args[0])
	},
}

var unregisterCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Unregister agent",
	Long:  `Unregister agent`,
	Run: func(cmd *cobra.Command, args []string) {
		unregisterAgent()
	},
}

var requestRegisterCmd = &cobra.Command{
	Use:   "request <access-token>",
	Short: "Request register from server",
	Long:  `Request register from server`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requestRegisterInfo(args[0])
	},
}

func init() {
	showRegisterCmd.Flags().BoolVar(&showSensitive, "sensitive", showSensitive, "show sensitive info")
	registerCmd.AddCommand(showRegisterCmd)
	registerCmd.AddCommand(setRegisterCmd)

	requestRegisterCmd.Flags().StringVar(&serverAddr, "server", serverAddr, "server address")
	requestRegisterCmd.Flags().StringVar(&uuid, "uuid", uuid, "agent uuid")
	requestRegisterCmd.Flags().StringVar(&deviceName, "name", deviceName, "agent name")
	registerCmd.AddCommand(requestRegisterCmd)
	registerCmd.AddCommand(unregisterCmd)
}

func getAgentInfo() (*metadata.AgentInfo, error) {
	conf, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, err
	}
	metadataDB, err := metadata.NewMetadataDB(conf.DataDir(), logrus.WithField("component", "metadataDB"))
	if err != nil {
		return nil, err
	}
	defer metadataDB.Close()

	info, err := metadataDB.GetAgentInfo()
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, nil
	}
	return info, nil
}

func showRegisterInfo() {
	info, err := getAgentInfo()
	if err != nil {
		logrus.WithError(err).Fatalf("get agent info")
		return
	}
	if info == nil {
		logrus.Fatalf("agent info is nil")
		return
	}
	if !showSensitive {
		info.Desensitization()
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		logrus.WithError(err).Fatalf("marshal agent info to json")
		return
	}
	fmt.Println(string(data))
}

func setAgentInfo(info *metadata.AgentInfo) error {
	conf, err := config.LoadConfig(configFile)
	if err != nil {
		logrus.Fatal("initConfig error, ", err.Error())
	}

	metadataDB, err := metadata.NewMetadataDB(conf.DataDir(), logrus.WithField("component", "metadataDB"))
	if err != nil {
		return err
	}
	defer metadataDB.Close()

	if err := metadataDB.UpdateAgentInfo(info); err != nil {
		return err
	}
	return nil
}

func setRegisterInfo(agentInfoPath string) {
	logrus.Infof("set register info from file: %s", agentInfoPath)
	data, err := os.ReadFile(agentInfoPath)
	if err != nil {
		logrus.WithError(err).Fatalf("read agent info file")
		return
	}

	var agentInfo metadata.AgentInfo
	if err2 := json.Unmarshal(data, &agentInfo); err2 != nil {
		logrus.WithError(err2).Fatalf("unmarshal agent info from file")
		return
	}
	registerTime := time.Now().Format(time.RFC3339)
	agentInfo.RegisterTime = &registerTime

	if err := setAgentInfo(&agentInfo); err != nil {
		logrus.WithError(err).Fatalf("set agent info")
		return
	}
	logrus.Infof("register agent %s success", *agentInfo.Uuid)
}

func requestRegisterInfo(accessToken string) {
	if deviceName == "" {
		hostname, _ := os.Hostname()
		deviceName = hostname
	}

	logrus.Infof("request register info from server: %s", serverAddr)
	req := dao.RegisterRequest{
		AccessToken: accessToken,
		Uuid:        uuid,
		Name:        deviceName,
	}
	jsonData, _ := json.Marshal(req)
	resp, err := http.Post(serverAddr+agentRegisterPath, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.WithError(err).Fatalf("request register info from server")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logrus.WithError(err).Fatalf("request register info from server, status code: %d", resp.StatusCode)
		return
	}
	var respBody dao.RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		logrus.WithError(err).Fatalf("decode register response")
		return
	}

	registerTime := time.Now().Format(time.RFC3339)
	agentInfo := &metadata.AgentInfo{
		Uuid:              &respBody.Uuid,
		Token:             &respBody.Token,
		S3AccessKeyID:     &respBody.S3AccessKeyID,
		S3SecretAccessKey: &respBody.S3SecretAccessKey,
		RegisterTime:      &registerTime,
	}

	if err := setAgentInfo(agentInfo); err != nil {
		logrus.WithError(err).Fatalf("set agent info")
		return
	}
	logrus.Infof("register agent %s success", *agentInfo.Uuid)
}

func unregisterAgent() {
	info, err := getAgentInfo()
	if err != nil {
		logrus.WithError(err).Fatalf("get agent info")
		return
	}
	if info == nil {
		logrus.Fatalf("agent info is nil")
		return
	}

	req, err := http.NewRequest(http.MethodPost, serverAddr+agentUnregisterPath, nil)
	if err != nil {
		logrus.WithError(err).Fatalf("new request")
		return
	}
	req.Header.Set("Authorization", "Bearer "+*info.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.WithError(err).Fatalf("unregister agent from server")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logrus.WithError(err).Fatalf("unregister agent from server, status code: %d", resp.StatusCode)
		return
	}
	logrus.Infof("unregister agent %s success", *info.Uuid)

	empty := ""
	info.Token = &empty
	info.S3AccessKeyID = &empty
	info.S3SecretAccessKey = &empty
	info.RegisterTime = &empty
	if err := setAgentInfo(info); err != nil {
		logrus.WithError(err).Fatalf("set agent info")
		return
	}
}
