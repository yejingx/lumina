package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"lumina/internal/config"
	"lumina/internal/model"
)

var insertTestData bool

var updateDBCommand = &cobra.Command{
	Use:   "updatedb",
	Short: "Update database tables",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.InitConfig(configFile)
		if err != nil {
			logrus.Fatal("initConfig error, ", err.Error())
		}

		db, err := model.InitDB(conf.DB)
		if err != nil {
			logrus.Fatal("failed to init database", err)
		}
		defer func() {
			sqlDb, _ := db.DB()
			sqlDb.Close()
		}()

		err = model.AutoMigrate(db)
		if err != nil {
			logrus.Fatal("failed to auto migrate database", err)
		} else {
			logrus.Infof("Database tables update successfully")
		}

		if insertTestData {
			err = model.InsertTestData(db)
			if err != nil {
				logrus.Fatal("failed to insert test data", err)
			}
		}
	},
}

func init() {
	updateDBCommand.Flags().BoolVarP(&insertTestData, "insert-test-data", "t", false, "Insert test data")
}
