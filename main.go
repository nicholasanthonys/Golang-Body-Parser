package main

import (
	"encoding/json"
	"fmt"
	"github.com/clbanning/mxj/x2j"
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

func middle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(e echo.Context) error {
		logrus.Info("method is ", e.Request().Method)
		return e.JSON(200, e.Request().Method)
		return next(e)
	}
}
func main() {

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//e.Use(middle)

	// Routes
	e.POST("/", switcher)
	e.PUT("/", switcher)
	e.GET("/", switcher)

	// Start server
	e.Logger.Fatal(e.Start(":8888"))
}

func getListFolder(dirname string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fmt.Println(file.Name())
	}
	return files, nil

}

func readConfigure(path string) []byte {
	// Open our jsonFile
	jsonFile, err := os.Open(path)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)
	return byteValue

}

func errorWriter(c echo.Context, configure model.Configure, err error, status int) error {
	responseMap := make(map[string]interface{})
	responseMap["message"] = err.Error()
	switch configure.Response.Transform {
	case "ToXml":
		logrus.Warn(err.Error())
		xmlByte, _ := x2j.MapToXml(responseMap)
		return c.XMLBlob(status, xmlByte)

	default:
		logrus.Warn(err.Error())
		return c.JSON(status, responseMap)
	}
}

//* Function that transform request to mpa[string] interface{}, Read configure J SON and return value
func switcher(c echo.Context) error {
	var arrRes []map[string]interface{}

	files, err := getListFolder("./configures")
	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In Reading File. " + err.Error()
		return c.JSON(http.StatusInternalServerError, resMap)
	}

	//*Read file Configure
	var configures []model.Configure
	var arrResByte [][]byte
	for _, file := range files {
		var configure model.Configure
		configByte := readConfigure("./configures/" + file.Name())
		//* assign configure byte to configure
		_ = json.Unmarshal(configByte, &configure)
		configures = append(configures, configure)

		_, resultMap := process(configure, c, arrRes)

		//*append to arr map strign inter
		arrRes = append(arrRes, resultMap)

		resByte, _ := service.TransformMapToByte(configure, resultMap)

		arrResByte = append(arrResByte, resByte)
	}

	//*response terakhir
	return service.ResponseWriter(configures[len(configures)-1], arrRes[len(arrRes)-1], c)
}

func process(configure model.Configure, c echo.Context, arrRes []map[string]interface{}) (int, map[string]interface{}) {

	//*this variable accept request from user
	requestFromUser := model.Fields{
		Header: make(map[string]interface{}),
		Body:   make(map[string]interface{}),
		Query:  make(map[string]interface{}),
	}
	resMap := make(map[string]interface{})
	//*check the content type user request
	contentType := c.Request().Header["Content-Type"][0]

	switch contentType {
	case "application/json":

		//*transform JSON request user to map request from user
		reqByte, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			logrus.Warn("error read request byte Json")
			resMap["message"] = err.Error()
			return http.StatusInternalServerError, resMap
		}
		requestFromUser.Body, err = service.FromJson(reqByte)

		if err != nil {
			logrus.Warn("error service from Json")
			resMap["message"] = err.Error()
			return http.StatusInternalServerError, resMap
		}

	case "application/x-www-form-urlencoded":
		//*transform x www form url encoded request user to map request from user
		requestFromUser.Body = service.FromFormUrl(c)
	case "application/xml":
		//*transform xml request user to map request from user
		reqByte, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			logrus.Warn("error read request byte xml")
			resMap["message"] = err.Error()
			return http.StatusInternalServerError, resMap
		}
		requestFromUser.Body, err = service.FromXmL(reqByte)
		if err != nil {
			logrus.Warn("error service from xml")
			resMap["message"] = err.Error()
			return http.StatusInternalServerError, resMap
		} else {
			logrus.Warn("service from xml success, request from user is")
			logrus.Warn(requestFromUser)
		}

	default:
		logrus.Warn("Content type not supported")
		resMap["message"] = "Content type not supported"
		return http.StatusBadRequest, resMap
	}

	//*get header value
	for key, val := range c.Request().Header {

		requestFromUser.Header[key] = val
	}

	//*get query value
	for key, val := range c.QueryParams() {
		requestFromUser.Query[key] = val
	}
	_, find := service.Find(configure.Methods, configure.Request.MethodUsed)
	if find {
		service.DoCommand(configure.Request, requestFromUser, arrRes)
	}

	//*send to destination url
	response, err := service.Send(configure, requestFromUser, configure.Request.MethodUsed, arrRes)

	if err != nil {
		//* return internal server error if there are any errors
		resMap["message"] = err.Error()
		return http.StatusInternalServerError, resMap
	} else {
		//* if there are no response from destination url, return a message
		if response == nil {
			resMap["message"] = "No response returned from destination url server"
			return http.StatusOK, resMap
		}
	}

	return http.StatusOK, response

}
