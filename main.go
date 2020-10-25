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
	"strconv"
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

	// Routes serial execution
	e.POST("/serial", doSerial)
	e.PUT("/serial", doSerial)
	e.GET("/serial", doSerial)

	// Routes parallel execution
	e.POST("/parallel", doParallel)
	e.PUT("/parallel", doParallel)
	e.GET("/parallel", doParallel)

	// Start server
	e.Logger.Fatal(e.Start(":8888"))
}

func worker(id int, wg *sync.WaitGroup, configure model.Configure, c echo.Context, arrRes []map[string]interface{}, mapRes map[string]interface{}, requestBody []byte) {
	defer wg.Done()
	fmt.Println("worker for id  ", id)
	_, resultMap := process(configure, c, arrRes, requestBody)
	//temp := make(map[string]interface{})
	//temp["configure"+strconv.Itoa(id)] = resultMap
	arrRes[id] = resultMap
	mapRes["configure"+strconv.Itoa(id)] = resultMap

	logrus.Info("worker ", id, " done")
}

func doParallel(c echo.Context) error {
	var wg sync.WaitGroup
	requestBody, _ := ioutil.ReadAll(c.Request().Body)
	files, err := service.GetListFolder("./configures")
	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In Reading File. " + err.Error()
		return c.JSON(http.StatusInternalServerError, resMap)
	}

	//*Read file Configure
	var configures []model.Configure

	arrRes := make([]map[string]interface{}, 10)
	mapRes := make(map[string]interface{})

	for index, file := range files {
		var configure model.Configure
		if strings.Contains(file.Name(), "configure") {
			configByte := service.ReadConfigure("./configures/" + file.Name())
			//* assign configure byte to configure
			_ = json.Unmarshal(configByte, &configure)
			configures = append(configures, configure)

			wg.Add(1)
			go worker(index, &wg, configure, c, arrRes, mapRes, requestBody)

		}

	}
	wg.Wait()

	var parallelConfig model.ParallelConfigure
	parallelConfigByte := service.ReadConfigure("./configures/parallel.json")

	_ = json.Unmarshal(parallelConfigByte, &parallelConfig)
	index := parallelConfig.ConfigureIndex

	//*use the latest configures and the latest response
	return service.ResponseWriter(configures[index], arrRes[index], c)
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

	//*Read file Configure
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
	requestFromUser := model.Fields{
		Header: make(map[string]interface{}),
		Body:   make(map[string]interface{}),
		Query:  make(map[string]interface{}),
	}
	resMap := make(map[string]interface{})
	//*check the content type user request
	contentType := c.Request().Header["Content-Type"][0]
	var err error
	switch contentType {
	case "application/json":

		//*transform JSON request user to map request from user
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
		logrus.Info("do modification")
		//* Do the Map Modification if method is find/available
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
