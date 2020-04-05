package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qshuai/coindis/controllers"
	"github.com/qshuai/coindis/models"
	"github.com/qshuai/coindis/routers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	viper.AddConfigPath("./")
	viper.SetConfigType("yaml")
	viper.SetConfigName("app")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("read configuration file error: %s\n", err)
		os.Exit(1)
	}

	level, err := logrus.ParseLevel(viper.GetString("log.level"))
	if err != nil {
		fmt.Printf("log.level configuration malformated: %s\n", err)
		os.Exit(1)
	}
	logrus.SetLevel(level)

	_, err = os.Stat("./logs/")
	if os.IsNotExist(err) {
		err = os.MkdirAll("./logs", 0766)
		if err != nil {
			fmt.Printf("create log directory error: %s\n", err)
			os.Exit(1)
		}
	}

	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05", DisableTimestamp: false})
	file, err := os.OpenFile("./logs/"+viper.GetString("log.filename")+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0644))
	if err != nil {
		fmt.Printf("open the specifed log file failed: %s\n", err)
		os.Exit(1)
	}
	logrus.SetOutput(file)

	// sync database state
	err = models.SyncDataSource()
	if err != nil {
		fmt.Printf("sync database failed: %s\n", err)
		os.Exit(1)
	}

	// init balance
	err = controllers.UpdateBalance()
	if err != nil {
		fmt.Printf("initial account balance error: %s\n", err)
		os.Exit(1)
	}

	go func() {
		ticker := time.NewTicker(time.Minute * 30)
		defer func() {
			ticker.Stop()
			logrus.Error("the goroutine holds the ticker to update balance existed")
		}()

		for _ = range ticker.C {
			err = controllers.UpdateBalance()
			if err != nil {
				logrus.Errorf("update account balance failed, please check: %s\n", err)
			}
		}
	}()

	controllers.InitCache()
	go func() {
		ticker := time.NewTicker(time.Minute * 30)
		defer func() {
			ticker.Stop()
			logrus.Error("the goroutine holds the ticker to cache clean existed")
		}()

		for _ = range ticker.C {
			controllers.CleanCache()
		}
	}()

	logrus.Info("the coindis program started!")

	r := gin.Default()
	r.LoadHTMLGlob("views/*")
	r.Static("/static", "./static")
	routers.RegisterApi(r)
	err = r.Run(":" + viper.GetString("app.port"))
	if err != nil {
		fmt.Printf("HTTP server runs error: %s\n", err)
	}
}
