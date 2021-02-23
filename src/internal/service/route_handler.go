package service

import (
	"encoding/json"
	"errors"
	"github.com/diegoholiveira/jsonlogic"
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
var projectByte []byte
var project model.Project
var fullProjectDirectory string
var logValue interface{} // value to be logged

//SetRouteHandler called by main.go. This function set route based on router.json
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
					e.POST(route.Path, doParallel, prepareRouteProject)
				} else {
					e.POST(route.Path, doSerial, prepareRouteProject)
				}
			}

			if strings.ToLower(route.Method) == "get" {
				if strings.ToLower(route.Type) == "parallel" {
					e.GET(route.Path, doParallel, prepareRouteProject)
				} else {
					e.GET(route.Path, doSerial, prepareRouteProject)
				}
			}

			if strings.ToLower(route.Method) == "put" {
				if strings.ToLower(route.Type) == "parallel" {
					e.PUT(route.Path, doParallel, prepareRouteProject)
				} else {
					e.PUT(route.Path, doSerial, prepareRouteProject)
				}

			}

			if strings.ToLower(route.Method) == "delete" {
				if strings.ToLower(route.Type) == "parallel" {
					e.DELETE(route.Path, doParallel, prepareRouteProject)
				} else {
					e.DELETE(route.Path, doSerial, prepareRouteProject)
				}
			}

		}
	}

	return e
}

// worker will called processingRequest. This function is called by doParallel function.
func worker(wg *sync.WaitGroup, mapKeyName string, configure model.Configure, c echo.Context, mapWrapper map[string]model.Wrapper, requestFromUser model.Wrapper, requestBody []byte) {
	defer wg.Done()
	_, status, err := processingRequest(mapKeyName, c, &requestFromUser, mapWrapper, requestBody)
	if err != nil {
		logrus.Error("Go Worker - Error Process")
		logrus.Error(err.Error())
		logrus.Error("status : ", status)
	}
	mapWrapper[mapKeyName] = requestFromUser
}

// prepareRouteProject middleware that find defined route in route.json and read project.json
func prepareRouteProject(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		index := util.FindRouteIndex(routes, c.Path())
		if index < 0 {
			return c.JSON(404, "Cannot Find Route "+c.Path())
		}
		route := routes[index]
		fullProjectDirectory = configureDir + "/" + route.ProjectDirectory
		logrus.Info("full project directory is")
		logrus.Info(fullProjectDirectory)

		// Read project .json
		projectByte = util.ReadJsonFile(fullProjectDirectory + "/" + "project.json")
		err := json.Unmarshal(projectByte, &project)

		if err != nil {
			resMap := make(map[string]string)
			resMap["message"] = "Problem In unmarshaling File project.json. "
			resMap["error"] = err.Error()
			return c.JSON(http.StatusInternalServerError, resMap)
		}

		return next(c)
	}

}

// doParallel execute every configure in parallel-way.
func doParallel(c echo.Context) error {
	// disable parallel for a while...
	return nil
	////*read the request that will be sent from user
	//reqByte, err := ioutil.ReadAll(c.Request().Body)
	//
	//if err != nil {
	//	resMap := make(map[string]string)
	//	resMap["message"] = "Problem In Reading Request Body. " + err.Error()
	//	return c.JSON(http.StatusInternalServerError, resMap)
	//}
	//
	////* declare a WaitGroup
	//var wg sync.WaitGroup
	//mapWrapper := make(map[string]model.Wrapper)
	//
	//for _, configureItem := range project.Configures {
	//	var configure model.Configure
	//	requestFromUser := model.Wrapper{
	//		Configure: configure,
	//		Request: model.Fields{
	//			Param:  make(map[string]interface{}),
	//			Header: make(map[string]interface{}),
	//			Body:   make(map[string]interface{}),
	//			Query:  make(map[string]interface{}),
	//		},
	//		Response: model.Fields{
	//			Param:  make(map[string]interface{}),
	//			Header: make(map[string]interface{}),
	//			Body:   make(map[string]interface{}),
	//			Query:  make(map[string]interface{}),
	//		},
	//	}
	//	configByte := util.ReadJsonFile(fullProjectDirectory + "/" + configureItem.FileName)
	//	//* assign configure byte to configure
	//	_ = json.Unmarshal(configByte, &configure)
	//	requestFromUser.Configure = configure
	//
	//	wg.Add(1)
	//	go worker(&wg, configureItem.Alias, configure, c, mapWrapper, requestFromUser, reqByte)
	//
	//}
	//wg.Wait()
	//
	////*now we need to parse the response.json command
	//resultWrapper := parseResponse(mapWrapper, fullProjectDirectory+"/response.json")
	//return util.ResponseWriter(resultWrapper, c)
}

// parseResponse process response (add,modify,delete) and return map to be sent to the client
func parseResponse(mapWrapper map[string]model.Wrapper, command model.Command) model.Wrapper {

	resultWrapper := model.Wrapper{
		Configure: model.Configure{
			Response: command,
		},
		Request: model.Fields{},
		Response: model.Fields{
			Param:  make(map[string]interface{}),
			Header: make(map[string]interface{}),
			Body:   make(map[string]interface{}),
			Query:  make(map[string]interface{}),
		},
	}

	//* now we will set the response body based from configurex.json if there is $configure value in configureBased.
	keyConfigure := util.RemoveCharacters(resultWrapper.Configure.ConfigureBased, "$")
	if strings.HasPrefix(resultWrapper.Configure.ConfigureBased, "$configure") {

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
	if len(resultWrapper.Configure.Response.LogAfterModify) > 0 {
		logValue = RetrieveValue(resultWrapper.Configure.Response.LogAfterModify, resultWrapper.Response)
		util.DoLogging(logValue, "after", "final response", false)
	}

	var statusCode int
	if resultWrapper.Configure.Response.StatusCode == 0 {
		// default
		statusCode = 400
	} else {
		statusCode = resultWrapper.Configure.Response.StatusCode
	}

	resultWrapper.Response.StatusCode = statusCode

	return resultWrapper
}

// doSerial process configure in serial-way.
func doSerial(c echo.Context) error {

	reqByte, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In Reading Request Body. " + err.Error()
		return c.JSON(http.StatusInternalServerError, resMap)
	}

	//*Read file ConfigureBased
	var mapWrapper = make(map[string]model.Wrapper) ///*slice that contains wrapper
	var mapConfigures = make(map[string]model.ConfigureItem)
	for _, configureItem := range project.Configures {
		//read actual configure based on configureItem.file_name
		// Initialization configure object
		var configure = model.Configure{}
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
		configByte := util.ReadJsonFile(fullProjectDirectory + "/" + configureItem.FileName)

		//* assign configure byte to configure
		_ = json.Unmarshal(configByte, &configure)
		requestFromUser.Configure = configure

		// store map alias - configure so it is easier to refer
		mapConfigures[configureItem.Alias] = configureItem

		// store map wrapper
		mapWrapper[configureItem.Alias] = requestFromUser

	}

	nextSuccess := project.Configures[0].CLogics[0].NextSuccess
	alias := project.Configures[0].Alias
	finalResponseConfigure := model.Command{}
	for len(strings.Trim(nextSuccess, " ")) > 0 {

		// Processing request
		requestFromUser := mapWrapper[alias]
		_, status, err := processingRequest(alias, c, &requestFromUser, mapWrapper, reqByte)
		if err != nil {
			return util.ErrorWriter(c, requestFromUser.Configure, err, status)
		}
		mapWrapper[alias] = requestFromUser

		var isAllLogicFail bool
		indexCLogic := 0
		for index, cLogicItem := range mapConfigures[alias].CLogics {
			indexCLogic = index
			InterfaceDirectModifier(cLogicItem.Rule, mapWrapper, "--")
			InterfaceDirectModifier(cLogicItem.Data, mapWrapper, "--")

			result, err := jsonlogic.ApplyInterface(cLogicItem.Rule, cLogicItem.Data)

			if err != nil {
				// break from loop to execute next failure
				break
			}

			if result.(bool) {
				// update next_sucess
				nextSuccess = cLogicItem.NextSuccess
				// update alias
				if len(strings.Trim(nextSuccess, " ")) > 0 {
					alias = nextSuccess
				}
				isAllLogicFail = false
				break
			} else {
				isAllLogicFail = true
			}

		}

		if isAllLogicFail {
			resultWrapper := parseResponse(mapWrapper, mapConfigures[alias].CLogics[indexCLogic].NextFailure)
			setHeaderResponse(resultWrapper.Response.Header, c)
			return util.ResponseWriter(resultWrapper, c)
		}

		if len(strings.Trim(nextSuccess, " ")) == 0 {
			finalResponseConfigure = mapConfigures[alias].CLogics[indexCLogic].Response
		}

	}

	resultWrapper := parseResponse(mapWrapper, finalResponseConfigure)
	setHeaderResponse(resultWrapper.Response.Header, c)
	return util.ResponseWriter(resultWrapper, c)

}

// setHeaderResponse set custom key-value pair for header, except Content-Length and Content-type
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

// processingRequest is the core function to process every configure. doCommand for transformation, send and receive request happen here.
func processingRequest(aliasName string, c echo.Context, wrapper *model.Wrapper, mapWrapper map[string]model.Wrapper, reqByte []byte) (*model.Wrapper, int, error) {

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
	if len(wrapper.Configure.Request.LogBeforeModify) > 0 {
		logValue = RetrieveValue(wrapper.Configure.Request.LogBeforeModify, wrapper.Request)
		util.DoLogging(logValue, "before", aliasName, true)
	}

	//*assign first before do any add,modification,delete in case value want reference each other
	mapWrapper[aliasName] = *wrapper

	//* Do the Map Modification
	DoCommand(wrapper.Configure.Request, wrapper.Request, mapWrapper)

	//*get the destinationPath value before sending request
	wrapper.Configure.Request.DestinationPath = ModifyPath(wrapper.Configure.Request.DestinationPath, "--", mapWrapper)

	//* In case user want to log after modify/changing request
	if len(wrapper.Configure.Request.LogAfterModify) > 0 {
		logValue = RetrieveValue(wrapper.Configure.Request.LogAfterModify, wrapper.Request)
		util.DoLogging(logValue, "after", aliasName, true)
	}

	//*send to destination url
	response, err := Send(wrapper)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	//*Modify responseByte in Receiver and get  byte from response that has been modified
	_, err = Receiver(wrapper.Configure, response, &wrapper.Response)

	//*close http
	defer response.Body.Close()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	//* In case user want to log before modify/changing request
	if len(wrapper.Configure.Response.LogBeforeModify) > 0 {
		logValue = RetrieveValue(wrapper.Configure.Response.LogBeforeModify, wrapper.Response)
		util.DoLogging(wrapper.Configure.Response.LogBeforeModify, "before", aliasName, false)
	}

	//* Do Command Add, Modify, Deletion for response again
	DoCommand(wrapper.Configure.Response, wrapper.Response, mapWrapper)

	//* In case user want to log after modify/changing request
	if len(wrapper.Configure.Response.LogAfterModify) > 0 {
		logValue = RetrieveValue(wrapper.Configure.Response.LogAfterModify, wrapper.Response)
		util.DoLogging(logValue, "after", aliasName, false)
	}
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
