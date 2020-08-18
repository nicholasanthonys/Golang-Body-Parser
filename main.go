package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/service"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
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
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/", Transform)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

func readConfigure(configure model.Configure) []byte {
	// Open our jsonFile
	jsonFile, err := os.Open("configure.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened configure.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)
	return byteValue

}

//* Function that transform request to mpa[string] interface{}, Read configure JSON and return value
func Transform(c echo.Context) error {
	//*Read file Configure
	var configure model.Configure
	configByte := readConfigure(configure)
	_ = json.Unmarshal(configByte, &configure)

	//*this variable accept request from user
	var requestFromUser map[string]interface{}

	//*check the content type
	contentType := c.Request().Header["Content-Type"][0]
	logrus.Info(contentType)
	switch contentType {
	case "application/json":
		myJson, _ := ioutil.ReadAll(c.Request().Body)
		json.Unmarshal(myJson, &requestFromUser)
		service.Add(configure, requestFromUser)
		service.Delete(configure, requestFromUser)
		service.Modify(configure, requestFromUser)
		service.Send(configure, requestFromUser)
		return c.JSON(http.StatusOK, requestFromUser)

	case "application/x-www-form-urlencoded":
		requestFromUser := service.FormUrlToMap(c)
		service.Add(configure, requestFromUser)
		service.Delete(configure, requestFromUser)
		service.Modify(configure, requestFromUser)

		service.Send(configure, requestFromUser)
		return c.JSON(http.StatusOK, requestFromUser)

	default:
		logrus.Info("Content type not supported")
		return c.JSON(http.StatusOK, "Type not supported")

	}

}
