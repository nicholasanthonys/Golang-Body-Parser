package request

import (
	"encoding/json"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/response"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	cmap "github.com/orcaman/concurrent-map"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

func DoParallel(cc *model.CustomContext, counter int) error {

	if counter == cc.BaseProject.MaxCircular {
		if &cc.BaseProject.CircularResponse != nil {
			resMap := response.ParseResponse(cc.MapWrapper, cc.BaseProject.CircularResponse, nil, nil)
			return response.ResponseWriter(resMap, cc.BaseProject.CircularResponse.Transform, cc)
		}
		resMap := make(map[string]interface{})
		resMap["message"] = "Circular Request detected"
		return cc.JSON(http.StatusLoopDetected, resMap)

	}

	// Read parallel.json
	ParallelProject := model.Parallel{}
	parallelByte := util.ReadJsonFile(cc.FullProjectDirectory + "/" + "parallel.json")
	err := json.Unmarshal(parallelByte, &ParallelProject)

	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In unmarshaling File parallel.json. "
		resMap["error"] = err.Error()
		return cc.JSON(http.StatusInternalServerError, resMap)
	}

	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In Reading Request Body. " + err.Error()
		return cc.JSON(http.StatusInternalServerError, resMap)
	}

	//* declare a WaitGroup
	var wg sync.WaitGroup

	var mapConfigures = make(map[string]model.ConfigureItem)

	for _, configureItem := range ParallelProject.Configures {
		var configure model.Configure
		requestFromUser := model.Wrapper{
			Configure: configure,
			Request:   cmap.New(),
			Response:  cmap.New(),
		}

		requestFromUser.Request.Set("param", make(map[string]interface{}))
		requestFromUser.Request.Set("header", make(map[string]interface{}))
		requestFromUser.Request.Set("body", make(map[string]interface{}))
		requestFromUser.Request.Set("query", make(map[string]interface{}))

		requestFromUser.Response.Set("statusCode", 0)
		requestFromUser.Response.Set("header", make(map[string]interface{}))
		requestFromUser.Response.Set("body", make(map[string]interface{}))

		configByte := util.ReadJsonFile(cc.FullProjectDirectory + "/" + configureItem.FileName)
		//* assign configure byte to configure
		_ = json.Unmarshal(configByte, &configure)
		requestFromUser.Configure = configure
		mapConfigures[configureItem.Alias] = configureItem

		loopIn := service.InterfaceDirectModifier(requestFromUser.Configure.Request.Loop, cc.MapWrapper, "--")
		lt := reflect.TypeOf(loopIn)
		var loop int
		if lt.Kind() == reflect.String {
			loop, err = strconv.Atoi(loopIn.(string))
			if err != nil {
				log.Error(err)
			}
		}

		if lt.Kind() == reflect.Int {
			loop = loopIn.(int)
		}

		if lt.Kind() == reflect.Float64 {
			loop = int(loopIn.(float64))
		}

		if loop == 0 {
			log.Info("set loop to 1 ")
			loop = 1
		}

		for i := 0; i < loop; i++ {
			if len(requestFromUser.Configure.Request.CLogics) > 0 {
				cLogicItem, _ := service.CLogicsChecker(requestFromUser.Configure.Request.CLogics, cc.MapWrapper)
				if cLogicItem != nil {
					wg.Add(1)
					go worker(&wg, configureItem.Alias, cc, &requestFromUser, i)
				}
			} else {
				// no clogics
				wg.Add(1)
				go worker(&wg, configureItem.Alias, cc, &requestFromUser, i)
			}

		}

	}
	wg.Wait()

	nextSuccess := ParallelProject.CLogics[0].NextSuccess
	finalResponseConfigure := model.Command{}
	for {

		cLogicItemTrue, err := service.CLogicsChecker(ParallelProject.CLogics, cc.MapWrapper)
		if err != nil {
			log.Error(err)
			tmpMapResponse := response.ParseResponse(cc.MapWrapper, ParallelProject.NextFailure, err, nil)

			return response.ResponseWriter(tmpMapResponse, ParallelProject.NextFailure.Transform, cc)
		}

		if cLogicItemTrue == nil {
			resultWrapper := response.ParseResponse(cc.MapWrapper, ParallelProject.NextFailure, nil, nil)
			response.SetHeaderResponse(resultWrapper.Header, cc)
			return response.ResponseWriter(resultWrapper, ParallelProject.NextFailure.Transform, cc)
		}

		// update next_success
		nextSuccess = cLogicItemTrue.NextSuccess
		// update alias
		if len(strings.Trim(nextSuccess, " ")) > 0 {
			if nextSuccess == "serial.json" {
				return DoSerial(cc, counter+1)
			}

			// reference to itself
			if nextSuccess == "parallel.json" {
				return DoParallel(cc, counter+1)
			}
		}
		if len(strings.Trim(nextSuccess, " ")) == 0 {
			finalResponseConfigure = cLogicItemTrue.Response
			break
		}

	}

	resultWrapper := response.ParseResponse(cc.MapWrapper, finalResponseConfigure, nil, nil)

	return response.ResponseWriter(resultWrapper, finalResponseConfigure.Transform, cc)

}

var mutex sync.Mutex

// worker will called ProcessingRequest. This function is called by parallelRouteHandler function.
func worker(wg *sync.WaitGroup, mapKeyName string, cc *model.CustomContext, requestFromUser *model.Wrapper, loopIndex int) {

	defer wg.Done()

	_, status, err := ProcessingRequest(mapKeyName, cc, requestFromUser, loopIndex)
	if err != nil {
		log.Error("Go Worker - Error Process")
		log.Error(err.Error())
		log.Error("status : ", status)
	}
	cc.MapWrapper.Set(mapKeyName, requestFromUser)
}
