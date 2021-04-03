package request

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/response"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

func DoParallel(c echo.Context, fullProjectDirectory string, mapWrapper map[string]model.Wrapper) error {

	// Read parallel.json
	ParallelProject := model.Parallel{}
	parallelByte := util.ReadJsonFile(fullProjectDirectory + "/" + "parallel.json")
	err := json.Unmarshal(parallelByte, &ParallelProject)

	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In unmarshaling File parallel.json. "
		resMap["error"] = err.Error()
		return c.JSON(http.StatusInternalServerError, resMap)
	}

	//*read the request that will be sent from user
	reqByte, err := ioutil.ReadAll(c.Request().Body)

	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In Reading Request Body. " + err.Error()
		return c.JSON(http.StatusInternalServerError, resMap)
	}

	//* declare a WaitGroup
	var wg sync.WaitGroup

	var mapConfigures = make(map[string]model.ConfigureItem)

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
		mapConfigures[configureItem.Alias] = configureItem

		loopIn := service.InterfaceDirectModifier(requestFromUser.Configure.Request.Loop, mapWrapper, "--")
		lt := reflect.TypeOf(loopIn)
		var loop int
		if lt.Kind() == reflect.String {
			loop, err = strconv.Atoi(loopIn.(string))
			if err != nil {
				log.Error(err)
				logrus.Info("set loop to 1 ")
			}
		}

		if lt.Kind() == reflect.Int {
			loop = loopIn.(int)
		}

		if lt.Kind() == reflect.Float64 {
			loop = int(loopIn.(float64))
		}

		for i := 0; i < loop; i++ {
			if len(requestFromUser.Configure.Request.CLogics) > 0 {
				cLogicItem, _ := service.CLogicsChecker(requestFromUser.Configure.Request.CLogics, mapWrapper)
				if cLogicItem != nil {
					wg.Add(1)
					go worker(&wg, configureItem.Alias, c, mapWrapper, requestFromUser, reqByte, i)
				}
			} else {
				// no clogics
				wg.Add(1)
				go worker(&wg, configureItem.Alias, c, mapWrapper, requestFromUser, reqByte, i)
			}

		}

	}
	wg.Wait()

	nextSuccess := ParallelProject.CLogics[0].NextSuccess
	finalResponseConfigure := model.Command{}
	for {

		cLogicItemTrue, err := service.CLogicsChecker(ParallelProject.CLogics, mapWrapper)
		if err != nil {
			logrus.Error(err)
			resultWrapper := response.ParseResponse(mapWrapper, ParallelProject.NextFailure)
			response.SetHeaderResponse(resultWrapper.Response.Header, c)
			return util.ResponseWriter(resultWrapper, c)
		}

		if cLogicItemTrue == nil {
			resultWrapper := response.ParseResponse(mapWrapper, ParallelProject.NextFailure)
			response.SetHeaderResponse(resultWrapper.Response.Header, c)
			return util.ResponseWriter(resultWrapper, c)
		}

		// update next_success
		nextSuccess = cLogicItemTrue.NextSuccess
		// update alias
		if len(strings.Trim(nextSuccess, " ")) > 0 {
			if nextSuccess == "serial.json" {
				return DoSerial(c, fullProjectDirectory, mapWrapper)
			}
		}
		if len(strings.Trim(nextSuccess, " ")) == 0 {
			finalResponseConfigure = cLogicItemTrue.Response
			break
		}

	}

	resultWrapper := response.ParseResponse(mapWrapper, finalResponseConfigure)
	response.SetHeaderResponse(resultWrapper.Response.Header, c)
	return util.ResponseWriter(resultWrapper, c)

	//var isAllLogicFail = false
	//var cLogicItemTrueIndex = 0

	//logrus.Info("parallel project")
	//logrus.Info(ParallelProject.CLogics)
	//// process the c logics
	//for index, cLogicItem := range ParallelProject.CLogics {
	//	cLogicItem.Rule = service.InterfaceDirectModifier(cLogicItem.Rule, mapWrapper, "--")
	//	cLogicItem.Data = service.InterfaceDirectModifier(cLogicItem.Data, mapWrapper, "--")
	//	result, err := jsonlogic.ApplyInterface(cLogicItem.Rule, cLogicItem.Data)
	//
	//	if err != nil {
	//		isAllLogicFail = true
	//		logrus.Error(err.Error())
	//		// send response
	//		resultWrapper := response.ParseResponse(mapWrapper, cLogicItem.Response)
	//		response.SetHeaderResponse(resultWrapper.Response.Header, c)
	//		return util.ResponseWriter(resultWrapper, c)
	//	}
	//	// get type of json logic result
	//	vt := reflect.TypeOf(result)
	//	if vt.Kind() == reflect.Bool {
	//		if result.(bool) {
	//			cLogicItemTrueIndex = index
	//			isAllLogicFail = false
	//			break
	//		} else {
	//			isAllLogicFail = true
	//		}
	//	}
	//}
	//
	//if !isAllLogicFail {
	//	resultWrapper := response.ParseResponse(mapWrapper, ParallelProject.CLogics[cLogicItemTrueIndex].Response)
	//	response.SetHeaderResponse(resultWrapper.Response.Header, c)
	//	return util.ResponseWriter(resultWrapper, c)
	//} else {
	//	resultWrapper := response.ParseResponse(mapWrapper, ParallelProject.NextFailure)
	//	response.SetHeaderResponse(resultWrapper.Response.Header, c)
	//	return util.ResponseWriter(resultWrapper, c)
	//}
}

// worker will called ProcessingRequest. This function is called by parallelRouteHandler function.
func worker(wg *sync.WaitGroup, mapKeyName string, c echo.Context, mapWrapper map[string]model.Wrapper, requestFromUser model.Wrapper, requestBody []byte, loopIndex int) {
	defer wg.Done()
	_, status, err := ProcessingRequest(mapKeyName, c, &requestFromUser, mapWrapper, requestBody, loopIndex)
	if err != nil {
		logrus.Error("Go Worker - Error Process")
		logrus.Error(err.Error())
		logrus.Error("status : ", status)
	}
	mapWrapper[mapKeyName] = requestFromUser
}
