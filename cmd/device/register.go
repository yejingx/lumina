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

	"lumina/internal/dao"
	"lumina/internal/device/config"
	"lumina/internal/device/metadata"
)

const (
	deviceRegisterPath   = "/api/v1/device/register"
	deviceUnregisterPath = "/api/v1/device/unregister"
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
	Short: "Show registered device info",
	Long:  `Show registered device info`,
	Run: func(cmd *cobra.Command, args []string) {
		showRegisterInfo()
	},
}

var setRegisterCmd = &cobra.Command{
	Use:   "set",
	Short: "Set registered device info",
	Long:  `Set registered device info`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setRegisterInfo(args[0])
	},
}

var unregisterCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Unregister device",
	Long:  `Unregister device`,
	Run: func(cmd *cobra.Command, args []string) {
		unregisterDevice()
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
	requestRegisterCmd.Flags().StringVar(&uuid, "uuid", uuid, "device uuid")
	requestRegisterCmd.Flags().StringVar(&deviceName, "name", deviceName, "device name")
	registerCmd.AddCommand(requestRegisterCmd)
	registerCmd.AddCommand(unregisterCmd)
}

func getDeviceInfo() (*metadata.DeviceInfo, error) {
	conf, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, err
	}
	metadataDB, err := metadata.NewMetadataDB(conf.DataDir(), logrus.WithField("component", "metadataDB"))
	if err != nil {
		return nil, err
	}
	defer metadataDB.Close()

	info, err := metadataDB.GetDeviceInfo()
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, nil
	}
	return info, nil
}

func showRegisterInfo() {
	info, err := getDeviceInfo()
	if err != nil {
		logrus.WithError(err).Fatalf("get device info")
		return
	}
	if info == nil {
		logrus.Fatalf("device info is nil")
		return
	}
	if !showSensitive {
		info.Desensitization()
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		logrus.WithError(err).Fatalf("marshal device info to json")
		return
	}
	fmt.Println(string(data))
}

func setDeviceInfo(info *metadata.DeviceInfo) error {
	conf, err := config.LoadConfig(configFile)
	if err != nil {
		logrus.Fatal("initConfig error, ", err.Error())
	}

	metadataDB, err := metadata.NewMetadataDB(conf.DataDir(), logrus.WithField("component", "metadataDB"))
	if err != nil {
		return err
	}
	defer metadataDB.Close()

	if err := metadataDB.UpdateDeviceInfo(info); err != nil {
		return err
	}
	return nil
}

func setRegisterInfo(deviceInfoPath string) {
	logrus.Infof("set register info from file: %s", deviceInfoPath)
	data, err := os.ReadFile(deviceInfoPath)
	if err != nil {
		logrus.WithError(err).Fatalf("read device info file")
		return
	}

	var deviceInfo metadata.DeviceInfo
	if err2 := json.Unmarshal(data, &deviceInfo); err2 != nil {
		logrus.WithError(err2).Fatalf("unmarshal device info from file")
		return
	}
	registerTime := time.Now().Format(time.RFC3339)
	deviceInfo.RegisterTime = &registerTime

	if err := setDeviceInfo(&deviceInfo); err != nil {
		logrus.WithError(err).Fatalf("set device info")
		return
	}
	logrus.Infof("register device %s success", *deviceInfo.Uuid)
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
	resp, err := http.Post(serverAddr+deviceRegisterPath, "application/json", bytes.NewBuffer(jsonData))
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
	deviceInfo := &metadata.DeviceInfo{
		Uuid:              &respBody.Uuid,
		Token:             &respBody.Token,
		S3AccessKeyID:     &respBody.S3AccessKeyID,
		S3SecretAccessKey: &respBody.S3SecretAccessKey,
		RegisterTime:      &registerTime,
	}

	if err := setDeviceInfo(deviceInfo); err != nil {
		logrus.WithError(err).Fatalf("set device info")
		return
	}
	logrus.Infof("register device %s success", *deviceInfo.Uuid)
}

func unregisterDevice() {
	info, err := getDeviceInfo()
	if err != nil {
		logrus.WithError(err).Fatalf("get device info")
		return
	}
	if info == nil {
		logrus.Fatalf("device info is nil")
		return
	}

	req, err := http.NewRequest(http.MethodPost, serverAddr+deviceUnregisterPath, nil)
	if err != nil {
		logrus.WithError(err).Fatalf("new request")
		return
	}
	req.Header.Set("Authorization", "Bearer "+*info.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.WithError(err).Fatalf("unregister device from server")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logrus.WithError(err).Fatalf("unregister device from server, status code: %d", resp.StatusCode)
		return
	}
	logrus.Infof("unregister device %s success", *info.Uuid)

	empty := ""
	info.Token = &empty
	info.S3AccessKeyID = &empty
	info.S3SecretAccessKey = &empty
	info.RegisterTime = &empty
	if err := setDeviceInfo(info); err != nil {
		logrus.WithError(err).Fatalf("set device info")
		return
	}
}
