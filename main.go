package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
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
		fmt.Printf("open log file error: %s\n", err)
		os.Exit(1)
	}
	logrus.SetOutput(file)

	// init balance
	updateBalance(&controller)

	controller.limit, err = controller.conf.Float("limit")
	if err != nil {
		panic(err)
	}

	// should justify whether the token configuration is empty or not
	controller.token = controller.conf.String("token")

	controller.interval, err = controller.conf.Int64("interval")
	if err != nil {
		panic(err)
	}

	go func() {
		ticker := time.NewTicker(time.Second * 30)
		defer func() {
			ticker.Stop()
			logrus.Error("the goroutine holds the ticker to update balance existed")
		}()

		for _ = range ticker.C {
			updateBalance(&controller)
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Minute * 10)
		defer func() {
			ticker.Stop()
			logrus.Error("the goroutine holds the ticker to cache clean existed")
		}()

		for _ = range ticker.C {
			controller.ic.Clean()
		}
	}()

	logrus.Info("the coindis program started!")

	r := gin.Default()
	routers.RegisterApi(r)

	err = r.Run("8080")
	if err != nil {
		logrus.Panicf("HTTP server runs error: %s", err)
	}
}
