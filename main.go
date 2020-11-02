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

func worker(wg *sync.WaitGroup, configure model.Configure, c echo.Context, mapWrapper map[string]model.Wrapper, requestFromUser model.Wrapper, requestBody []byte, fileName string) {
	defer wg.Done()
	process(configure, c, &requestFromUser, mapWrapper, requestBody)
	mapWrapper[fileName] = requestFromUser
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

	mapWrapper := make(map[string]model.Wrapper)

	for _, file := range files {
		var configure model.Configure
		if strings.Contains(file.Name(), "configure") {

			requestFromUser := model.Wrapper{
				Configure: configure,
				Request: model.Fields{
					Param:  make(map[string]interface{}),
					Header: make(map[string]interface{}),
					Body:   make(map[string]interface{}),
					Query:  make(map[string]interface{}),
				},
				Response: model.Fields{
					Param:  make(map[string]interface{}),
					Header: make(map[string]interface{}),
					Body:   make(map[string]interface{}),
					Query:  make(map[string]interface{}),
				},
			}
			configByte := service.ReadConfigure("./configures/" + file.Name())
			//* assign configure byte to configure
			_ = json.Unmarshal(configByte, &configure)
			requestFromUser.Configure = configure

			wg.Add(1)
			go worker(&wg, configure, c, mapWrapper, requestFromUser, requestBody, file.Name())

		}

	}
	wg.Wait()

	//*read configure response.json
	var parallelResponse model.Configure
	parallelConfigByte := service.ReadConfigure("./configures/response.json")
	_ = json.Unmarshal(parallelConfigByte, &parallelResponse)

	//*now we need to parse the response.json command
	resultWrapper := parseResponseParallel(mapWrapper, parallelResponse)
	return service.ResponseWriter(resultWrapper, c)
}

func parseResponseParallel(mapWrapper map[string]model.Wrapper, parallelResponse model.Configure) model.Wrapper {
	//var resultMap = make(map[string]interface{})
	//fields := model.Fields{
	//	Param:  make(map[string]interface{}),
	//	Header: make(map[string]interface{}),
	//	Body:   make(map[string]interface{}),
	//	Query:  make(map[string]interface{}),
	//}

	resultWrapper := model.Wrapper{
		Configure: model.Configure{},
		Request:   model.Fields{},
		Response: model.Fields{
			Param:  make(map[string]interface{}),
			Header: make(map[string]interface{}),
			Body:   make(map[string]interface{}),
			Query:  make(map[string]interface{}),
		},
	}

	for key, value := range parallelResponse.Response.Adds.Body {
		stringValue := fmt.Sprintf("%v", value)
		if strings.HasPrefix(stringValue, "$configure") {
			//// * split between $configure-$request-value
			//valueSplice := strings.Split(stringValue, "-")
			////* get the traverse key
			//listTraverseKey := strings.Split(key, ".")
			////* sanitized value from square bracket and dollar sign
			//sanitizedValue, _ := service.SanitizeValue(fmt.Sprintf("%v", valueSplice[1]))
			////* get the real value
			//realValue := service.GetValue(sanitizedValue, mapWrapper[valueSplice[0]].Response.Body, 0)
			/////* add recursive key-value
			////service.AddRecursive(listTraverseKey, fmt.Sprintf("%v", realValue), resultMap, 0)
			service.DoCommandConfigureBody(parallelResponse.Response, resultWrapper.Response, mapWrapper)
		} else {
			resultWrapper.Response.Body[key] = value
		}
	}
	return resultWrapper
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
	//var configures []model.Configure                //* slice for configures file (JSON)
	var mapWrapper = make(map[string]model.Wrapper) ///*slice that contains wrapper

	for _, file := range files {
		var configure model.Configure
		if strings.Contains(file.Name(), "configure") {

			//* Make a wrapper for each configuration
			requestFromUser := model.Wrapper{
				Configure: configure,
				Request: model.Fields{
					Param:  make(map[string]interface{}),
					Header: make(map[string]interface{}),
					Body:   make(map[string]interface{}),
					Query:  make(map[string]interface{}),
				},
				Response: model.Fields{
					Param:  make(map[string]interface{}),
					Header: make(map[string]interface{}),
					Body:   make(map[string]interface{}),
					Query:  make(map[string]interface{}),
				},
			}

			configByte := service.ReadConfigure("./configures/" + file.Name())
			//* assign configure byte to configure
			_ = json.Unmarshal(configByte, &configure)
			requestFromUser.Configure = configure

			process(configure, c, &requestFromUser, mapWrapper, reqByte)

			//*store to temporary map
			//*append to arr map string model wrapper
			mapWrapper[file.Name()] = requestFromUser

		}

	}

	//*use the latest configures and the latest response
	//return c.JSON(200, mapWrapper["configure1.json"].Response.Body)
	return service.ResponseWriter(mapWrapper["configure1.json"], c)
}

func process(configure model.Configure, c echo.Context, wrapperUser *model.Wrapper, mapWrapper map[string]model.Wrapper, reqByte []byte) (int, *model.Wrapper) {

	//*this variable accept request from user

	resMap := make(map[string]interface{})
	//*check the content type user request
	contentType := c.Request().Header["Content-Type"][0]
	var err error
	switch contentType {
	case "application/json":

		//*transform JSON request user to map request from user
		wrapperUser.Request.Body, err = service.FromJson(reqByte)
		if err != nil {
			logrus.Warn("error service from Json")
			wrapperUser.Response.Body["message"] = err.Error()
			return http.StatusInternalServerError, wrapperUser
		}

	case "application/x-www-form-urlencoded":
		//*transform x www form url encoded request user to map request from user
		wrapperUser.Request.Body = service.FromFormUrl(c)
	case "application/xml":

		//if err != nil {
		//	logrus.Warn("error read request byte xml")
		//	wrapperUser.Response.Body["message"] = err.Error()
		//	return http.StatusInternalServerError, nil
		//}
		//*transform xml request user to map request from user
		wrapperUser.Request.Body, err = service.FromXmL(reqByte)
		if err != nil {
			logrus.Warn("error service from xml")
			wrapperUser.Response.Body["message"] = err.Error()
			return http.StatusInternalServerError, nil
		} else {
			logrus.Warn("service from xml success, request from user is")
			logrus.Warn(wrapperUser.Request)
		}

	default:
		logrus.Warn("Content type not supported")
		resMap["message"] = "Content type not supported"
		return http.StatusBadRequest, wrapperUser
	}

	//*set header value
	for key, _ := range c.Request().Header {

		wrapperUser.Request.Header[key] = c.Request().Header.Get(key)
	}

	//*set query value
	for key, _ := range c.QueryParams() {
		wrapperUser.Request.Query[key] = c.QueryParam(key)
	}

	//*set param value
	for _, value := range c.ParamNames() {
		wrapperUser.Request.Param[value] = c.Param(value)
	}

	_, find := service.Find(configure.Methods, configure.Request.MethodUsed)
	if find {
		//* Do the Map Modification if method is find/available
		service.DoCommand(configure.Request, wrapperUser.Request, mapWrapper)
	}

	//*send to destination url
	response, err := service.Send(configure, wrapperUser, configure.Request.MethodUsed)
	//*Modifty responseByte in Receiver and get  byte from response that has been modified
	_, err = service.Receiver(configure, response, &wrapperUser.Response)
	defer response.Body.Close()

	if err != nil {
		return 0, nil
	}

	//* response always do Command
	service.DoCommand(configure.Response, wrapperUser.Response, mapWrapper)

	//if err != nil {
	//	//* return internal server error if there are any errors
	//	wrapperUser.Response.Body["message"] = err.Error()
	//	return http.StatusInternalServerError, nil
	//} else {
	//	//* if there are no response from destination url, return a message
	//	if response == nil {
	//		wrapperUser.Response.Body["message"] = "No response returned from destination url server"
	//		return http.StatusOK, nil
	//	}
	//}

	return http.StatusOK, wrapperUser

}
