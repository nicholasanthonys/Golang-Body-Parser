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
	"strings"
	"sync"
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

	files, _ := service.GetListFolder("./configures")

	//*set path based from configure
	for _, file := range files {
		var configure model.Configure
		if strings.Contains(file.Name(), "configure") {
			configByte := service.ReadConfigure("./configures/" + file.Name())
			//* assign configure byte to configure
			_ = json.Unmarshal(configByte, &configure)
			// Routes serial execution
			e.POST("/serial"+configure.Path, doSerial)
			e.PUT("/serial"+configure.Path, doSerial)
			e.GET("/serial"+configure.Path, doSerial)
			// Routes parallel execution
			e.POST("/parallel"+configure.Path, doParallel)
			e.PUT("/parallel"+configure.Path, doParallel)
			e.GET("/parallel"+configure.Path, doParallel)
		}
	}

	// Start server
	e.Logger.Fatal(e.Start(":8888"))
}

func worker(wg *sync.WaitGroup, configure model.Configure, c echo.Context, arrRes *[]map[string]interface{}, mapRes map[string]interface{}, requestBody []byte, fileName string) {
	defer wg.Done()
	_, resultMap := process(configure, c, *arrRes, requestBody)
	*arrRes = append(*arrRes, resultMap)
	mapRes[fileName] = resultMap
}

func doParallel(c echo.Context) error {
	//* declare a WaitGroup
	var wg sync.WaitGroup

	//*read the request that will be sent from user
	requestBody, _ := ioutil.ReadAll(c.Request().Body)
	//* get files and store it in slice
	files, err := service.GetListFolder("./configures")
	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In Reading File. " + err.Error()
		return c.JSON(http.StatusInternalServerError, resMap)
	}

	var configures = make(map[string]model.Configure)
	var mapC = make(map[string]echo.Context)
	arrRes := make([]map[string]interface{}, 0)
	mapRes := make(map[string]interface{})

	for _, file := range files {
		var configure model.Configure
		if strings.Contains(file.Name(), "configure") {
			configByte := service.ReadConfigure("./configures/" + file.Name())
			//* assign configure byte to configure
			_ = json.Unmarshal(configByte, &configure)
			configures[file.Name()] = configure
			mapC[file.Name()] = c
			wg.Add(1)
			go worker(&wg, configure, c, &arrRes, mapRes, requestBody, file.Name())

		}

	}
	wg.Wait()

	var parallelResponse model.Configure
	parallelConfigByte := service.ReadConfigure("./configures/response.json")

	_ = json.Unmarshal(parallelConfigByte, &parallelResponse)
	parsedMap := parseResponseParallel(mapRes, parallelResponse, mapC)
	return service.ResponseWriter(parallelResponse, parsedMap, c)
}

func parseResponseParallel(mapRes map[string]interface{}, parallelResponse model.Configure, mapC map[string]echo.Context) map[string]interface{} {
	var resultMap = make(map[string]interface{})
	//var requestFromUser model.Fields
	//resultMap := make(map[string]interface{})
	for key, value := range parallelResponse.Response.Adds.Body {
		stringValue := fmt.Sprintf("%v", value)
		if strings.HasPrefix(stringValue, "configure") {

			valueSplice := strings.Split(stringValue, "-")
			parallelResponse.Response.Adds.Body[key] = valueSplice[1]
			logrus.Info("parallelResponse.Response.Adds.Body[key] is ", parallelResponse.Response.Adds.Body[key])
			logrus.Info("value splice1  is ", valueSplice[1])
			listTraverseKey := strings.Split(key, ".")
			sanitizedValue, _ := service.SanitizeValue(fmt.Sprintf("%v", valueSplice[1]))
			realValue := service.GetValue(sanitizedValue, mapRes[valueSplice[0]], 0)
			logrus.Info("realValue is ", realValue)
			service.AddRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), resultMap, 0)

		} else {
			resultMap[key] = value
		}

	}

	return resultMap

}

//* Function that transform request to mpa[string] interface{}, Read configure JSON and return value
func doSerial(c echo.Context) error {

	files, err := service.GetListFolder("./configures")
	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In Reading File. " + err.Error()
		return c.JSON(http.StatusInternalServerError, resMap)
	}

	reqByte, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In Reading Request Body. " + err.Error()
		return c.JSON(http.StatusInternalServerError, resMap)
	}

	//*Read file ConfigureBased
	var configures []model.Configure //* slice for configures file (JSON)

	var arrRes []map[string]interface{} ///*slice that contains response in map string interface

	for _, file := range files {
		var configure model.Configure
		if strings.Contains(file.Name(), "configure") {

			configByte := service.ReadConfigure("./configures/" + file.Name())
			//* assign configure byte to configure
			_ = json.Unmarshal(configByte, &configure)
			configures = append(configures, configure)

			_, resultMap := process(configure, c, arrRes, reqByte)

			//*append to arr map string inter
			arrRes = append(arrRes, resultMap)

		}

	}

	//*use the latest configures and the latest response
	return service.ResponseWriter(configures[len(configures)-1], arrRes[len(arrRes)-1], c)
}

func process(configure model.Configure, c echo.Context, arrRes []map[string]interface{}, reqByte []byte) (int, map[string]interface{}) {

	//*this variable accept request from user
	requestFromUser := model.Wrapper{
		Request: model.Fields{
			Header: make(map[string]interface{}),
			Body:   make(map[string]interface{}),
			Query:  make(map[string]interface{}),
		},
		Response: model.Fields{
			Header: make(map[string]interface{}),
			Body:   make(map[string]interface{}),
			Query:  make(map[string]interface{}),
		},
	}
	resMap := make(map[string]interface{})
	//*check the content type user request
	contentType := c.Request().Header["Content-Type"][0]
	var err error
	switch contentType {
	case "application/json":

		//*transform JSON request user to map request from user
		requestFromUser.Request.Body, err = service.FromJson(reqByte)
		if err != nil {
			logrus.Warn("error service from Json")
			resMap["message"] = err.Error()
			return http.StatusInternalServerError, resMap
		}

	case "application/x-www-form-urlencoded":
		//*transform x www form url encoded request user to map request from user
		requestFromUser.Request.Body = service.FromFormUrl(c)
	case "application/xml":
		//*transform xml request user to map request from user
		if err != nil {
			logrus.Warn("error read request byte xml")
			resMap["message"] = err.Error()
			return http.StatusInternalServerError, resMap
		}
		requestFromUser.Request.Body, err = service.FromXmL(reqByte)
		if err != nil {
			logrus.Warn("error service from xml")
			resMap["message"] = err.Error()
			return http.StatusInternalServerError, resMap
		} else {
			logrus.Warn("service from xml success, request from user is")
			logrus.Warn(requestFromUser.Request)
		}

	default:
		logrus.Warn("Content type not supported")
		resMap["message"] = "Content type not supported"
		return http.StatusBadRequest, resMap
	}

	//*get header value
	for key, val := range c.Request().Header {

		requestFromUser.Request.Header[key] = val
	}

	//*get query value
	for key, val := range c.QueryParams() {
		requestFromUser.Request.Query[key] = val
	}

	_, find := service.Find(configure.Methods, configure.Request.MethodUsed)
	if find {
		logrus.Info("do modification")
		//* Do the Map Modification if method is find/available
		service.DoCommand(c, configure.Request, requestFromUser.Request, arrRes)
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
