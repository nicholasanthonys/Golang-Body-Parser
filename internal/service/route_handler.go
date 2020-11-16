package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

func SetRouteHandler() *echo.Echo {

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//e.Use(middle)

	files, err := GetListFolder("./configures")

	if err != nil {
		logrus.Error("error reading directory ./configures")

	}

	//*set path based from configure
	for _, file := range files {
		var configure model.Configure
		if strings.Contains(file.Name(), "configure") {
			configByte := ReadConfigure("./configures/" + file.Name())
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

	return e

}

func worker(wg *sync.WaitGroup, fileName string, configure model.Configure, c echo.Context, mapWrapper map[string]model.Wrapper, requestFromUser model.Wrapper, requestBody []byte) {
	defer wg.Done()
	_, status, err := process(fileName, configure, c, &requestFromUser, mapWrapper, requestBody)
	if err != nil {
		logrus.Error("Go Worker - Error Process")
		logrus.Error(err.Error())
		logrus.Error("status : ", status)
	}
	mapWrapper[fileName] = requestFromUser
}

func doParallel(c echo.Context) error {

	//* declare a WaitGroup
	var wg sync.WaitGroup

	//*read the request that will be sent from user
	requestBody, _ := ioutil.ReadAll(c.Request().Body)
	//* get files and store it in slice
	files, err := GetListFolder("./configures")
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
			configByte := ReadConfigure("./configures/" + file.Name())
			//* assign configure byte to configure
			_ = json.Unmarshal(configByte, &configure)
			requestFromUser.Configure = configure

			wg.Add(1)
			go worker(&wg, file.Name(), configure, c, mapWrapper, requestFromUser, requestBody)

		}

	}
	wg.Wait()

	//*now we need to parse the response.json command
	resultWrapper := parseResponse(mapWrapper)
	return ResponseWriter(resultWrapper, c)
}

func parseResponse(mapWrapper map[string]model.Wrapper) model.Wrapper {

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

	parallelConfigByte := ReadConfigure("./configures/response.json")
	_ = json.Unmarshal(parallelConfigByte, &resultWrapper.Configure)

	//* meants that the response is based from configurex.json
	if strings.HasPrefix(resultWrapper.Configure.ConfigureBased, "$configure") {
		keyConfigure := RemoveCharacters(resultWrapper.Configure.ConfigureBased, "$")
		//*check if key exist in the map
		if _, ok := mapWrapper[keyConfigure]; ok {
			//* get configureX.json from map wrapper
			resultWrapper.Response = mapWrapper[keyConfigure].Response
		}
	}

	for key, value := range resultWrapper.Configure.Response.Adds.Body {
		stringValue := fmt.Sprintf("%v", value)
		if strings.HasPrefix(stringValue, "$configure") {
			//*add
			AddToWrapper(resultWrapper.Configure.Response.Adds.Body, "--", resultWrapper.Response.Body, mapWrapper)
			//*modify
			ModifyWrapper(resultWrapper.Configure.Response.Modifies.Body, "--", resultWrapper.Response.Body, mapWrapper)
			//* delete
			DeletionBody(resultWrapper.Configure.Response.Deletes, resultWrapper.Response)
		} else {
			resultWrapper.Response.Body[key] = value
		}
	}
	logrus.Info("result is ")
	logrus.Info(resultWrapper)
	return resultWrapper
}

//* Function that transform request to mpa[string] interface{}, Read configure JSON and return value
func doSerial(c echo.Context) error {

	files, err := GetListFolder("./configures")
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

			configByte := ReadConfigure("./configures/" + file.Name())
			//* assign configure byte to configure
			_ = json.Unmarshal(configByte, &configure)
			requestFromUser.Configure = configure

			_, status, err := process(file.Name(), configure, c, &requestFromUser, mapWrapper, reqByte)

			if err != nil {
				return ErrorWriter(c, configure, err, status)
			}

			//*save to map
			mapWrapper[file.Name()] = requestFromUser

		}

	}

	//*use the latest configures and the latest response
	//return c.JSON(200, mapWrapper["configure1.json"].Response.Body)
	resultWrapper := parseResponse(mapWrapper)
	return ResponseWriter(resultWrapper, c)
}

func process(fileName string, configure model.Configure, c echo.Context, wrapperUser *model.Wrapper, mapWrapper map[string]model.Wrapper, reqByte []byte) (*model.Wrapper, int, error) {

	//*this variable accept request from user

	//*check the content type user request
	var contentType string
	if c.Request().Header["Content-Type"] != nil {
		contentType = c.Request().Header["Content-Type"][0]
	} else {
		contentType = "application/json"
	}

	var err error
	switch contentType {
	case "application/json":
		logrus.Info("content type is json")
		//*transform JSON request user to map request from user
		wrapperUser.Request.Body, err = FromJson(reqByte)
		if err != nil {
			logrus.Warn("error service from Json")
			wrapperUser.Response.Body["message"] = err.Error()
			return nil, http.StatusInternalServerError, err
		}

	case "application/x-www-form-urlencoded":
		//*transform x www form url encoded request user to map request from user
		wrapperUser.Request.Body = FromFormUrl(c)
	case "application/xml":

		//*transform xml request user to map request from user
		wrapperUser.Request.Body, err = FromXmL(reqByte)
		if err != nil {
			logrus.Warn("error service from xml")
			wrapperUser.Response.Body["message"] = err.Error()
			return nil, http.StatusInternalServerError, err
		} else {
			logrus.Warn("service from xml success, request from user is")
			logrus.Warn(wrapperUser.Request)
		}

	default:
		logrus.Warn("Content type not supported")
		resMap := make(map[string]interface{})
		resMap["message"] = "Content type not supported"
		return nil, http.StatusBadRequest, errors.New("Content Type Not Supported")
	}

	//*set header value
	for key := range c.Request().Header {

		wrapperUser.Request.Header[key] = c.Request().Header.Get(key)
	}

	//*set query value
	for key := range c.QueryParams() {
		wrapperUser.Request.Query[key] = c.QueryParam(key)
	}

	//*set param value
	for _, value := range c.ParamNames() {
		wrapperUser.Request.Param[value] = c.Param(value)
	}

	_, find := Find(configure.Methods, configure.Request.MethodUsed)
	if find {
		//*assign first before do any add,modification,delete in case value want reference each other
		mapWrapper[fileName] = *wrapperUser
		//* Do the Map Modification if method is find/available
		DoCommand(configure.Request, wrapperUser.Request, mapWrapper)
	}

	//*get the desinationPath value before sending request
	configure.Request.DestinationPath = ModifyPath(configure.Request.DestinationPath, "--", mapWrapper)

	//*send to destination url
	response, err := Send(configure, wrapperUser, configure.Request.MethodUsed)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	//*Modify responseByte in Receiver and get  byte from response that has been modified
	_, err = Receiver(configure, response, &wrapperUser.Response)
	//*close http
	defer response.Body.Close()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	//* Do Command Add, Modify, Deletion again
	DoCommand(configure.Response, wrapperUser.Response, mapWrapper)

	return wrapperUser, http.StatusOK, nil

}
