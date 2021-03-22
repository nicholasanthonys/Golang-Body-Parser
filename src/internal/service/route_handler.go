package service

import (
	"encoding/json"
	"errors"
	"github.com/diegoholiveira/jsonlogic"
	"github.com/jinzhu/copier"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var configureDir string
var routes model.Routes
var serialByte []byte
var parallelByte []byte
var SerialProject model.Serial
var ParallelProject model.Parallel
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
					e.POST(route.Path, doParallel, prepareParallelRoute)
				} else {
					e.POST(route.Path,
						doSerial, prepareSerialRoute)
				}
			}

			if strings.ToLower(route.Method) == "get" {
				if strings.ToLower(route.Type) == "parallel" {
					e.GET(route.Path, doParallel, prepareParallelRoute)
				} else {
					e.GET(route.Path, doSerial, prepareSerialRoute)
				}
			}

			if strings.ToLower(route.Method) == "put" {
				if strings.ToLower(route.Type) == "parallel" {
					e.PUT(route.Path, doParallel, prepareParallelRoute)
				} else {
					e.PUT(route.Path, doSerial, prepareSerialRoute)
				}

			}

			if strings.ToLower(route.Method) == "delete" {
				if strings.ToLower(route.Type) == "parallel" {
					e.DELETE(route.Path, doParallel, prepareParallelRoute)
				} else {
					e.DELETE(route.Path, doSerial, prepareSerialRoute)
				}
			}

		}
	}

	return e
}

// worker will called processingRequest. This function is called by doParallel function.
func worker(wg *sync.WaitGroup, mapKeyName string, configure model.Configure, c echo.Context, mapWrapper map[string]model.Wrapper, requestFromUser model.Wrapper, requestBody []byte, loopIndex int) {
	defer wg.Done()
	_, status, err := processingRequest(mapKeyName, c, &requestFromUser, mapWrapper, requestBody, loopIndex)
	if err != nil {
		logrus.Error("Go Worker - Error Process")
		logrus.Error(err.Error())
		logrus.Error("status : ", status)
	}
	mapWrapper[mapKeyName] = requestFromUser
}

// prepareSerialRoute middleware that find defined route in route.json and read SerialProject.json
func prepareSerialRoute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		index := util.FindRouteIndex(routes, c.Path())
		if index < 0 {
			return c.JSON(404, "Cannot FindInSliceOfString Route "+c.Path())
		}
		route := routes[index]
		fullProjectDirectory = configureDir + "/" + route.ProjectDirectory
		logrus.Info("full SerialProject directory is")
		logrus.Info(fullProjectDirectory)

		SerialProject = model.Serial{}
		// Read SerialProject .json
		serialByte = util.ReadJsonFile(fullProjectDirectory + "/" + "serial.json")
		err := json.Unmarshal(serialByte, &SerialProject)

		if err != nil {
			resMap := make(map[string]string)
			resMap["message"] = "Problem In unmarshaling File serial.json. "
			resMap["error"] = err.Error()
			return c.JSON(http.StatusInternalServerError, resMap)
		}
		return next(c)
	}
}

// prepareSerialRoute middleware that find defined route in route.json and read SerialProject.json
func prepareParallelRoute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		index := util.FindRouteIndex(routes, c.Path())
		if index < 0 {
			return c.JSON(404, "Cannot FindInSliceOfString Route "+c.Path())
		}
		route := routes[index]
		fullProjectDirectory = configureDir + "/" + route.ProjectDirectory
		logrus.Info("full SerialProject directory is")
		logrus.Info(fullProjectDirectory)

		// Read parallel.json
		ParallelProject = model.Parallel{}
		parallelByte = util.ReadJsonFile(fullProjectDirectory + "/" + "parallel.json")
		err := json.Unmarshal(parallelByte, &ParallelProject)

		if err != nil {
			resMap := make(map[string]string)
			resMap["message"] = "Problem In unmarshaling File parallel.json. "
			resMap["error"] = err.Error()
			return c.JSON(http.StatusInternalServerError, resMap)
		}
		return next(c)
	}
}

// doParallel execute every configure in parallel-way.
func doParallel(c echo.Context) error {
	//*read the request that will be sent from user
	reqByte, err := ioutil.ReadAll(c.Request().Body)

	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In Reading Request Body. " + err.Error()
		return c.JSON(http.StatusInternalServerError, resMap)
	}

	//* declare a WaitGroup
	var wg sync.WaitGroup
	mapWrapper := make(map[string]model.Wrapper)

	for _, configureItem := range ParallelProject.Configures {
		var configure model.Configure
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

		loop := requestFromUser.Configure.Request.Loop
		if loop == 0 {
			loop = 1
		}
		for i := 0; i < loop; i++ {
			wg.Add(1)
			go worker(&wg, configureItem.Alias, configure, c, mapWrapper, requestFromUser, reqByte, i)
		}

	}
	wg.Wait()

	var isAllLogicFail = false
	var cLogicItemTrueIndex = 0
	// process the c logics
	for index, cLogicItem := range ParallelProject.CLogics {
		InterfaceDirectModifier(cLogicItem.Rule, mapWrapper, "--")
		InterfaceDirectModifier(cLogicItem.Data, mapWrapper, "--")
		result, err := jsonlogic.ApplyInterface(cLogicItem.Rule, cLogicItem.Data)

		if err != nil {
			isAllLogicFail = true
			logrus.Error(err.Error())
			// send response
			resultWrapper := parseResponse(mapWrapper, cLogicItem.Response)
			setHeaderResponse(resultWrapper.Response.Header, c)
			return util.ResponseWriter(resultWrapper, c)
		}
		// get type of json logic result
		vt := reflect.TypeOf(result)
		if vt.Kind() == reflect.Bool {
			if result.(bool) {
				cLogicItemTrueIndex = index
				isAllLogicFail = false
				break
			} else {
				isAllLogicFail = true
			}
		}
	}

	if !isAllLogicFail {
		resultWrapper := parseResponse(mapWrapper, ParallelProject.CLogics[cLogicItemTrueIndex].Response)
		setHeaderResponse(resultWrapper.Response.Header, c)
		return util.ResponseWriter(resultWrapper, c)
	} else {
		resultWrapper := parseResponse(mapWrapper, ParallelProject.NextFailure)
		setHeaderResponse(resultWrapper.Response.Header, c)
		return util.ResponseWriter(resultWrapper, c)
	}

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
	//keyConfigure := util.RemoveCharacters(resultWrapper.Configure.ConfigureBased, "$")
	//if strings.HasPrefix(resultWrapper.Configure.ConfigureBased, "$configure") {
	//
	//	//*check if key exist in the map
	//	if _, ok := mapWrapper[keyConfigure]; ok {
	//		//* get configureX.json from map wrapper
	//		resultWrapper.Response = mapWrapper[keyConfigure].Response
	//	}
	//}

	//*header
	AddToWrapper(resultWrapper.Configure.Response.Adds.Header, "--", resultWrapper.Response.Header, mapWrapper, 0)
	//*modify header
	ModifyWrapper(resultWrapper.Configure.Response.Modifies.Header, "--", resultWrapper.Response.Header, mapWrapper, 0)
	//*Deletion Header
	DeletionHeaderOrQuery(resultWrapper.Configure.Response.Deletes.Header, resultWrapper.Response.Header)

	//*add
	AddToWrapper(resultWrapper.Configure.Response.Adds.Body, "--", resultWrapper.Response.Body, mapWrapper, 0)
	//*modify
	ModifyWrapper(resultWrapper.Configure.Response.Modifies.Body, "--", resultWrapper.Response.Body, mapWrapper, 0)
	//* delete
	DeletionBody(resultWrapper.Configure.Response.Deletes, resultWrapper.Response)

	//*In case user want to log final response
	if len(resultWrapper.Configure.Response.LogAfterModify) > 0 {
		logValue = RetrieveValue(resultWrapper.Configure.Response.LogAfterModify, resultWrapper.Response, 0)
		util.DoLogging(logValue, "after", "final response", false)
	}

	var statusCode int
	if resultWrapper.Configure.Response.StatusCode == 0 {
		// default
		statusCode = 400
	} else {
		statusCode = resultWrapper.Configure.Response.StatusCode
	}

	resultWrapper.Response.StatusCode = strconv.Itoa(statusCode)

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

	for _, configureItem := range SerialProject.Configures {
		//read actual configure based on configureItem.file_name
		// Initialization configure object
		var configure = model.Configure{}
		requestFromUser := model.Wrapper{
			Configure: model.Configure{},
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

	// assumption :  the first configure to be processed is configures at index 0 from SerialProject.configures
	nextSuccess := SerialProject.Configures[0].CLogics[0].NextSuccess
	alias := SerialProject.Configures[0].Alias
	finalResponseConfigure := model.Command{}

	for len(strings.Trim(nextSuccess, " ")) > 0 {

		// Processing request
		requestFromUser := mapWrapper[alias]
		// Loop only available for parallel request, therefore, set loopIndex to 0
		_, _, err := processingRequest(alias, c, &requestFromUser, mapWrapper, reqByte, 0)

		if err != nil {
			// next failure
			resultWrapper := parseResponse(mapWrapper, mapConfigures[alias].NextFailure)
			setHeaderResponse(resultWrapper.Response.Header, c)
			return util.ResponseWriter(resultWrapper, c)
			//return util.ErrorWriter(c, requestFromUser.Configure, err, status)

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
				logrus.Error("error is ")
				logrus.Error(err.Error())
				// break from loop to execute next failure
				break
			}

			// get type of json logic result
			vt := reflect.TypeOf(result)
			if vt.Kind() == reflect.Bool {
				if result.(bool) {
					isAllLogicFail = false
					nextSuccess = cLogicItem.NextSuccess
					// update alias
					if len(strings.Trim(nextSuccess, " ")) > 0 {
						alias = nextSuccess
					}
					break
				} else {
					isAllLogicFail = true
					break
				}
			}

			// update next_sucess
			nextSuccess = cLogicItem.NextSuccess
			// update alias
			if len(strings.Trim(nextSuccess, " ")) > 0 {
				alias = nextSuccess
			}

		}

		if isAllLogicFail {
			resultWrapper := parseResponse(mapWrapper, mapConfigures[alias].NextFailure)
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
func processingRequest(aliasName string, c echo.Context, wrapper *model.Wrapper, mapWrapper map[string]model.Wrapper, reqByte []byte, loopIndex int) (*model.Wrapper, int, error) {
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
		logValue = RetrieveValue(wrapper.Configure.Request.LogBeforeModify, wrapper.Request, loopIndex)
		util.DoLogging(logValue, "before", aliasName, true)
	}

	//*assign first before do any add,modification,delete in case value want reference each other
	mapWrapper[aliasName] = *wrapper

	// copy wrapper
	tempWrapper := model.Wrapper{}
	copier.Copy(&tempWrapper, &wrapper)

	//* Do the Map Modification
	DoCommand(tempWrapper.Configure.Request, tempWrapper.Request, mapWrapper, loopIndex)

	//*get the destinationPath value before sending request
	tempWrapper.Configure.Request.DestinationPath = ModifyPath(tempWrapper.Configure.Request.DestinationPath, "--", mapWrapper, loopIndex)

	//* In case user want to log after modify/changing request
	if len(tempWrapper.Configure.Request.LogAfterModify) > 0 {
		logValue = RetrieveValue(tempWrapper.Configure.Request.LogAfterModify, tempWrapper.Request, loopIndex)
		util.DoLogging(logValue, "after", aliasName, true)
	}

	//*send to destination url
	response, err := Send(wrapper)

	if err != nil {
		logrus.Error("Error send : ", err.Error())
		return nil, http.StatusInternalServerError, err
	}

	//*Modify responseByte in Receiver and get  byte from response that has been modified
	_, err = Receiver(tempWrapper.Configure, response, &wrapper.Response)

	//*close http
	defer response.Body.Close()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	//* In case user want to log before modify/changing request
	if len(tempWrapper.Configure.Response.LogBeforeModify) > 0 {
		logValue = RetrieveValue(tempWrapper.Configure.Response.LogBeforeModify, tempWrapper.Response, loopIndex)
		util.DoLogging(tempWrapper.Configure.Response.LogBeforeModify, "before", aliasName, false)
	}

	//* Do Command Add, Modify, Deletion for response again
	DoCommand(tempWrapper.Configure.Response, tempWrapper.Response, mapWrapper, loopIndex)

	//* In case user want to log after modify/changing request
	if len(tempWrapper.Configure.Response.LogAfterModify) > 0 {
		logValue = RetrieveValue(tempWrapper.Configure.Response.LogAfterModify, tempWrapper.Response, loopIndex)
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
