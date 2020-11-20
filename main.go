package main

import (
	"github.com/joho/godotenv"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/sirupsen/logrus"
	"os"
)

var log = logrus.New()

func init() {
	//* init logger with timestamp
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	log.Level = logrus.DebugLevel
}

func main() {
	//*load env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//*start Server
	err = service.SetRouteHandler().Start(os.Getenv("APP_URL") + ":" + os.Getenv("APP_PORT"))
	if err != nil {
		panic(err)
	}
}
