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

var myMap map[string]interface{}

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

func readConfigure() []byte {
	// Open our jsonFile
	jsonFile, err := os.Open("configure.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	logrus.Info("Successfully Opened configure.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)
	return byteValue

}

//* Function that transform request to mpa[string] interface{}, Read configure JSON and return value
func switcher(c echo.Context) error {
	//*Read file Configure
	var configure model.Configure
	configByte := readConfigure()

	//* assign configure byte to configure
	_ = json.Unmarshal(configByte, &configure)

	//*this variable accept request from user
	requestFromUser := model.Fields{
		Header: make(map[string]interface{}),
		Body:   make(map[string]interface{}),
		Query:  make(map[string]interface{}),
	}

	//*check the content type user request
	contentType := c.Request().Header["Content-Type"][0]
	logrus.Info(contentType)

	switch contentType {
	case "application/json":

		//*transform JSON request user to map request from user
		reqByte, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			logrus.Warn("error read request byte Json")
			logrus.Warn(err.Error())
			os.Exit(1)
		}
		requestFromUser.Body, err = service.FromJson(reqByte)
		if err != nil {
			logrus.Warn("error service from Json")
			logrus.Warn(err.Error())
			os.Exit(1)
		}

	case "application/x-www-form-urlencoded":
		//*transform x www form url encoded request user to map request from user
		requestFromUser.Body = service.FromFormUrl(c)
	case "application/xml":
		//*transform xml request user to map request from user
		reqByte, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			logrus.Warn("error read request byte xml")
			logrus.Warn(err.Error())
			os.Exit(1)
		}
		requestFromUser.Body, err = service.FromXmL(reqByte)
		if err != nil {
			logrus.Warn("error service from xml")
			os.Exit(1)
		} else {
			logrus.Warn("service from xml success, request from user is")
			logrus.Warn(requestFromUser)
		}

	default:
		logrus.Info("Content type not supported")
		return c.JSON(http.StatusOK, "Type not supported")
	}

	//*do map modification for request
	service.DoCommandConfigureBody(configure.Request, requestFromUser)

	//*send to destination url
	response, err := service.Send(configure, requestFromUser)

	if err != nil {
		//* return internal server error if there are any errors
		return c.JSON(http.StatusInternalServerError, err.Error())
	} else {
		//* if there are no response from destination url, return a message
		if response == nil {
			defaultResponse := make(map[string]string)
			defaultResponse["message"] = "No response returned from destination url server"
			return c.JSON(http.StatusOK, defaultResponse)
		}
	}

	//*return response
	switch configure.Response.Transform {
	case "ToJson":
		return c.JSONBlob(http.StatusOK, response)
	case "ToXml":
		return c.XMLBlob(http.StatusOK, response)
	default:
		defaultResponse := make(map[string]interface{})
		defaultResponse["message"] = string(response)
		return c.JSON(http.StatusBadRequest, defaultResponse)
	}

}
