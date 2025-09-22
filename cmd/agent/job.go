package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"lumina/internal/agent/config"
	"lumina/internal/agent/metadata"
	"lumina/internal/dao"
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Job management tools",
	Long:  `Manage and monitor jobs running on the agent`,
}

var addJobCmd = &cobra.Command{
	Use:   "add <path to job.json>",
	Short: "Add a job to the agent",
	Long:  `Add a job to the agent for processing`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		addJob(args[0])
	},
}

var listJobCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all jobs",
	Long:    `List all jobs stored in the agent`,
	Run: func(cmd *cobra.Command, args []string) {
		listJobs()
	},
}

var deleteJobCmd = &cobra.Command{
	Use:     "delete <job-uuid>",
	Aliases: []string{"rm", "del"},
	Short:   "Delete a job",
	Long:    `Delete a job from the agent by UUID`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deleteJob(args[0])
	},
}

func init() {
	jobCmd.AddCommand(addJobCmd)
	jobCmd.AddCommand(listJobCmd)
	jobCmd.AddCommand(deleteJobCmd)
}

func addJob(jobFilePath string) {
	logrus.Infof("Adding job from file: %s", jobFilePath)
	conf, err := config.LoadConfig(configFile)
	if err != nil {
		logrus.Fatal("initConfig error, ", err.Error())
	}

	// read job from json file
	data, err := os.ReadFile(jobFilePath)
	if err != nil {
		logrus.WithError(err).Fatalf("read job from file")
		return
	}

	var job dao.JobSpec
	if err2 := json.Unmarshal(data, &job); err2 != nil {
		logrus.WithError(err2).Fatalf("unmarshal job from file")
		return
	}

	metadataDB, err := metadata.NewMetadataDB(conf.DataDir(), logrus.WithField("component", "metadataDB"))
	if err != nil {
		logrus.WithError(err).Fatalf("new metadata db")
		return
	}
	defer metadataDB.Close()

	if err := metadataDB.SetJob(job.Uuid, &job); err != nil {
		logrus.WithError(err).Fatalf("set job to metadata db")
		return
	}

	logrus.Infof("add job %s success", job.Uuid)
}

func listJobs() {
	logrus.Info("Listing all jobs")
	conf, err := config.LoadConfig(configFile)
	if err != nil {
		logrus.Fatal("initConfig error, ", err.Error())
	}

	metadataDB, err := metadata.NewMetadataDB(conf.DataDir(), logrus.WithField("component", "metadataDB"))
	if err != nil {
		logrus.WithError(err).Fatalf("new metadata db")
		return
	}
	defer metadataDB.Close()

	jobs, err := metadataDB.GetJobs()
	if err != nil {
		logrus.WithError(err).Fatalf("get jobs from metadata db")
		return
	}

	if len(jobs) == 0 {
		logrus.Info("No jobs found")
		return
	}

	for i, job := range jobs {
		jobJSON, err := json.MarshalIndent(job, "", "  ")
		if err != nil {
			logrus.WithError(err).Errorf("marshal job %s to JSON", job.Uuid)
			continue
		}
		fmt.Printf("Job %d: %s\n", i, string(jobJSON))
	}
}

func deleteJob(jobUuid string) {
	logrus.Infof("Deleting job: %s", jobUuid)
	conf, err := config.LoadConfig(configFile)
	if err != nil {
		logrus.Fatal("initConfig error, ", err.Error())
	}

	metadataDB, err := metadata.NewMetadataDB(conf.DataDir(), logrus.WithField("component", "metadataDB"))
	if err != nil {
		logrus.WithError(err).Fatalf("new metadata db")
		return
	}
	defer metadataDB.Close()

	// Check if job exists
	job, err := metadataDB.GetJob(jobUuid)
	if err != nil {
		logrus.WithError(err).Fatalf("get job from metadata db")
		return
	}
	if job == nil {
		logrus.Errorf("job %s not found", jobUuid)
		return
	}

	if err := metadataDB.DeleteJob(jobUuid); err != nil {
		logrus.WithError(err).Fatalf("delete job from metadata db")
		return
	}

	logrus.Infof("delete job %s success", jobUuid)
}
