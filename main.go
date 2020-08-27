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
	"plugin"
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

var toXml func(map[string]interface{}) []byte

func main() {

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/", switcher)

	// Start server
	e.Logger.Fatal(e.Start(":8888"))
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

func Load(c echo.Context) error {
	plug, err := plugin.Open("./plugin/transform.so")
	if err != nil {
		logrus.Warn("Unable to load plugin module")
		logrus.Warn(err.Error())
		c.JSON(http.StatusInternalServerError, err)
		os.Exit(1)

	}
	_, err = plug.Lookup("ToJson")
	if err != nil {
		logrus.Warn("Unable to Lookup plugin module")
		logrus.Warn(err.Error())
		c.JSON(http.StatusInternalServerError, err)
		os.Exit(1)
	}
	return c.JSON(http.StatusOK, "load succeess")
}

//* Function that transform request to mpa[string] interface{}, Read configure JSON and return value
func switcher(c echo.Context) error {
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
		requestFromUser, _ = service.FromJson(myJson)

	case "application/x-www-form-urlencoded":
		requestFromUser = service.FromFormUrl(c)

	default:
		logrus.Info("Content type not supported")
		return c.JSON(http.StatusOK, "Type not supported")
	}

	//*do map modification
	ModifyMap(configure, requestFromUser)

	//*send
	response, err := service.Send(configure, requestFromUser)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	switch configure.Response.Transform {
	case "ToXml":
		return c.XMLBlob(http.StatusOK, response)
	}
	return c.JSONBlob(http.StatusOK, response)
}

func ModifyMap(configure model.Configure, requestFromUser map[string]interface{}) {
	service.Add(configure, requestFromUser)
	service.Delete(configure, requestFromUser)
	service.Modify(configure, requestFromUser)
}
