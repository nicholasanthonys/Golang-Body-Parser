package service

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
)

var configureDir string
var routes model.Routes

func SetRouteHandler() *echo.Echo {
	//* get configures Directory
	configureDir = os.Getenv("CONFIGURES_DIRECTORY")

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//e.Use(middle)

	// * Read router.json
	routesByte := util.ReadJsonFile(configureDir + "/router.json")
	err := json.Unmarshal(routesByte, &routes)
	if err != nil {
		logrus.Error(err.Error())
	} else {
		//*add index route
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Golang-Body-Parser Active")
		})

		//*set path based from configure
		for _, route := range routes {

			if strings.ToLower(route.Method) == "post" {
				if strings.ToLower(route.Type) == "parallel" {
					e.POST(route.Path, doParallel)
				} else {
					e.POST(route.Path, doSerial)
				}
			}

			if strings.ToLower(route.Method) == "get" {
				if strings.ToLower(route.Type) == "parallel" {
					e.GET(route.Path, doParallel)
				} else {
					e.GET(route.Path, doSerial)
				}
			}

			if strings.ToLower(route.Method) == "put" {
				if strings.ToLower(route.Type) == "parallel" {
					e.PUT(route.Path, doParallel)
				} else {
					e.PUT(route.Path, doSerial)
				}

			}

			if strings.ToLower(route.Method) == "delete" {
				if strings.ToLower(route.Type) == "parallel" {
					e.DELETE(route.Path, doParallel)
				} else {
					e.DELETE(route.Path, doSerial)
				}
			}

			////* assign configure byte to configure
			//_ = json.Unmarshal(configByte, &configure)
			//// Route serial execution
			//e.POST("/serial"+configure.Path, doSerial)
			//e.PUT("/serial"+configure.Path, doSerial)
			//e.GET("/serial"+configure.Path, doSerial)
			//// Route parallel execution
			//e.POST("/parallel"+configure.Path, doParallel)
			//e.PUT("/parallel"+configure.Path, doParallel)
			//e.GET("/parallel"+configure.Path, doParallel)

		}
	}

	//files, err := util.GetListFolder(configureDir)
	//
	//if err != nil {
	//	logrus.Fatal("error reading directory " + configureDir)
	//
	//}

	return e
}

func worker(wg *sync.WaitGroup, fileName string, configure model.Configure, c echo.Context, mapWrapper map[string]model.Wrapper, requestFromUser model.Wrapper, requestBody []byte) {
	defer wg.Done()
	_, status, err := processingRequest(fileName, configure, c, &requestFromUser, mapWrapper, requestBody)
	if err != nil {
		logrus.Error("Go Worker - Error Process")
		logrus.Error(err.Error())
		logrus.Error("status : ", status)
	}
	mapWrapper[fileName] = requestFromUser
}

func doParallel(c echo.Context) error {
	index := util.FindRouteIndex(routes, c.Path())
	if index < 0 {
		return c.JSON(404, "Cannot Find Route "+c.Path())
	}
	route := routes[index]
	fullProjectDirectory := configureDir + "/" + route.ProjectDirectory

	//* declare a WaitGroup
	var wg sync.WaitGroup

	//*read the request that will be sent from user
	requestBody, _ := ioutil.ReadAll(c.Request().Body)
	//* get files and store it in slice
	files, err := util.GetListFolder(fullProjectDirectory)
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
			configByte := util.ReadJsonFile(fullProjectDirectory + "/" + file.Name())
			//* assign configure byte to configure
			_ = json.Unmarshal(configByte, &configure)
			requestFromUser.Configure = configure

			wg.Add(1)
			go worker(&wg, file.Name(), configure, c, mapWrapper, requestFromUser, requestBody)

		}

	}
	wg.Wait()

	//*now we need to parse the response.json command
	resultWrapper := parseResponse(mapWrapper, fullProjectDirectory+"/response.json")
	return util.ResponseWriter(resultWrapper, c)
}

func parseResponse(mapWrapper map[string]model.Wrapper, responsePath string) model.Wrapper {

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

	parallelConfigByte := util.ReadJsonFile(responsePath)
	_ = json.Unmarshal(parallelConfigByte, &resultWrapper.Configure)

	//* now we will set the response body based from configurex.json if there is $configure value in configureBased.
	if strings.HasPrefix(resultWrapper.Configure.ConfigureBased, "$configure") {
		keyConfigure := util.RemoveCharacters(resultWrapper.Configure.ConfigureBased, "$")
		//*check if key exist in the map
		if _, ok := mapWrapper[keyConfigure]; ok {
			//* get configureX.json from map wrapper
			resultWrapper.Response = mapWrapper[keyConfigure].Response
		}
	}

	//*header
	AddToWrapper(resultWrapper.Configure.Response.Adds.Header, "--", resultWrapper.Response.Header, mapWrapper)
	//*modify header
	ModifyWrapper(resultWrapper.Configure.Response.Modifies.Header, "--", resultWrapper.Response.Header, mapWrapper)
	//*Deletion Header
	DeletionHeaderOrQuery(resultWrapper.Configure.Response.Deletes.Header, resultWrapper.Response.Header)

	//*add
	AddToWrapper(resultWrapper.Configure.Response.Adds.Body, "--", resultWrapper.Response.Body, mapWrapper)
	//*modify
	ModifyWrapper(resultWrapper.Configure.Response.Modifies.Body, "--", resultWrapper.Response.Body, mapWrapper)
	//* delete
	DeletionBody(resultWrapper.Configure.Response.Deletes, resultWrapper.Response)

	//*In case user want to log final response
	DoLogging(resultWrapper.Configure.Response.LogAfterModify, resultWrapper.Response, "after", "final response", false)
	return resultWrapper
}

//* Function that transform request to mpa[string] interface{}, Read configure JSON and return value
func doSerial(c echo.Context) error {
	index := util.FindRouteIndex(routes, c.Path())
	if index < 0 {
		return c.JSON(404, "Cannot Find Route "+c.Path())
	}

	route := routes[index]
	fullProjectDirectory := configureDir + "/" + route.ProjectDirectory
	files, err := util.GetListFolder(fullProjectDirectory)

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

			configByte := util.ReadJsonFile(fullProjectDirectory + "/" + file.Name())
			//* assign configure byte to configure
			_ = json.Unmarshal(configByte, &configure)
			requestFromUser.Configure = configure

			_, status, err := processingRequest(file.Name(), configure, c, &requestFromUser, mapWrapper, reqByte)

			if err != nil {
				return util.ErrorWriter(c, configure, err, status)
			}

			//*save to map
			mapWrapper[file.Name()] = requestFromUser

		}

	}

	//*use the latest configures and the latest response
	//return c.JSON(200, mapWrapper["configure1.json"].Response.Body)
	resultWrapper := parseResponse(mapWrapper, fullProjectDirectory+"/response.json")
	//* for each value in map wrapper header, set the header
	setHeaderResponse(resultWrapper.Response.Header, c)
	return util.ResponseWriter(resultWrapper, c)
}

func setHeaderResponse(header map[string]interface{}, c echo.Context) {
	for key, val := range header {
		rt := reflect.TypeOf(val)
		//* only add if interface type is string
		if rt.Kind() == reflect.String {
			if key != "Content-Length" && key != "Content-Type" {
				c.Response().Header().Set(key, val.(string))
			}

		}
	}

}

func processingRequest(fileName string, configure model.Configure, c echo.Context, wrapper *model.Wrapper, mapWrapper map[string]model.Wrapper, reqByte []byte) (*model.Wrapper, int, error) {

	//*check the content type user request
	var contentType string
	var err error
	var status int

	if c.Request().Header["Content-Type"] != nil {
		contentType = c.Request().Header["Content-Type"][0]
	} else {
		contentType = "application/json"
	}

	//*convert request to map string interface based on the content type
	wrapper.Request.Body, status, err = parseRequestBody(c, contentType, reqByte)
	if err != nil {
		return nil, status, err
	}

	//*set header value
	for key := range c.Request().Header {

		wrapper.Request.Header[key] = c.Request().Header.Get(key)
	}

	//*set query value
	for key := range c.QueryParams() {
		wrapper.Request.Query[key] = c.QueryParam(key)
	}

	//*set param value
	for _, value := range c.ParamNames() {

		wrapper.Request.Param[value] = c.Param(value)
	}

	//* In case user want to log before modify/changing request
	DoLogging(configure.Request.LogBeforeModify, wrapper.Request, "before", fileName, true)

	//*if methodUsed is in the array of configure methods, then do the map modification
	//_, find := util.Find(configure.Methods, configure.Request.Method)
	//if find {
	//
	//}

	//*assign first before do any add,modification,delete in case value want reference each other
	mapWrapper[fileName] = *wrapper
	//* Do the Map Modification if method is find/available
	DoCommand(configure.Request, wrapper.Request, mapWrapper)

	//*get the destinationPath value before sending request
	configure.Request.DestinationPath = ModifyPath(configure.Request.DestinationPath, "--", mapWrapper)

	//* In case user want to log after modify/changing request
	DoLogging(configure.Request.LogAfterModify, wrapper.Request, "after", fileName, true)

	//*send to destination url
	response, err := Send(configure, wrapper, configure.Request.Method)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	//*Modify responseByte in Receiver and get  byte from response that has been modified
	_, err = Receiver(configure, response, &wrapper.Response)
	//*close http
	defer response.Body.Close()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	//* In case user want to log before modify/changing request
	DoLogging(configure.Response.LogBeforeModify, wrapper.Response, "before", fileName, false)

	//* Do Command Add, Modify, Deletion for response again
	DoCommand(configure.Response, wrapper.Response, mapWrapper)

	//* In case user want to log before modify/changing request
	DoLogging(configure.Response.LogAfterModify, wrapper.Response, "after", fileName, false)

	return wrapper, http.StatusOK, nil

}

func parseRequestBody(c echo.Context, contentType string, reqByte []byte) (map[string]interface{}, int, error) {
	var err error
	var result = make(map[string]interface{})
	switch contentType {
	case "application/json":
		//*transform JSON request user to map request from user
		result, err = FromJson(reqByte)
		if err != nil {
			logrus.Warn("error parse request body from Json")
			result["message"] = err.Error()
			return nil, http.StatusInternalServerError, err
		}

	case "application/x-www-form-urlencoded":
		//*transform x www form url encoded request user to map request from user
		result = FromFormUrl(c)
	case "application/xml":

		//*transform xml request user to map request from user
		result, err = FromXmL(reqByte)
		if err != nil {
			logrus.Warn("error service from xml")
			result["message"] = err.Error()
			return nil, http.StatusInternalServerError, err
		}

	default:
		logrus.Warn("Content type not supported")
		return nil, http.StatusBadRequest, errors.New("Content Type Not Supported")
	}
	return result, http.StatusOK, nil
}

func DoLogging(logValue string, field model.Fields, event string, fileName string, isRequest bool) {
	if len(logValue) > 0 {
		sentence := "logging "
		if isRequest {
			sentence += "response "
		} else {
			sentence += "response "
		}

		if event == "before" {
			sentence += "before modify for " + fileName + " : "
		} else {
			sentence += "after modify for " + fileName + " : "
		}

		value := CheckValue(logValue, field)
		logrus.Info(sentence)
		logrus.Info(value)
	}
}
