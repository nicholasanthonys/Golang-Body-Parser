package request

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/response"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	cmap "github.com/orcaman/concurrent-map"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

func DoParallel(c echo.Context, baseProject model.Base, fullProjectDirectory string, mapWrapper cmap.ConcurrentMap, counter int) error {

	if counter == baseProject.MaxCircular {
		if &baseProject.CircularResponse != nil {
			resMap := response.ParseResponse(mapWrapper, baseProject.CircularResponse)
			return response.ResponseWriter(resMap, baseProject.CircularResponse.Transform, c)
		}
		resMap := make(map[string]interface{})
		resMap["message"] = "Circular Request detected"
		return c.JSON(http.StatusLoopDetected, resMap)

	}

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
			Request:   cmap.New(),
			Response:  cmap.New(),
		}

		requestFromUser.Request.Set("param", make(map[string]interface{}))
		requestFromUser.Request.Set("header", make(map[string]interface{}))
		requestFromUser.Request.Set("body", make(map[string]interface{}))
		requestFromUser.Request.Set("query", make(map[string]interface{}))

		requestFromUser.Response.Set("statusCode", "")
		requestFromUser.Response.Set("header", make(map[string]interface{}))
		requestFromUser.Response.Set("body", make(map[string]interface{}))

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
				log.Info("set loop to 1 ")
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
			log.Error(err)
			tmpMapResponse := response.ParseResponse(mapWrapper, ParallelProject.NextFailure)

			return response.ResponseWriter(tmpMapResponse, ParallelProject.NextFailure.Transform, c)
		}

		if cLogicItemTrue == nil {
			resultWrapper := response.ParseResponse(mapWrapper, ParallelProject.NextFailure)

			c = response.SetHeaderResponse(resultWrapper["header"].(map[string]interface{}), c)
			return response.ResponseWriter(resultWrapper, ParallelProject.NextFailure.Transform, c)
		}

		// update next_success
		nextSuccess = cLogicItemTrue.NextSuccess
		// update alias
		if len(strings.Trim(nextSuccess, " ")) > 0 {
			if nextSuccess == "serial.json" {
				return DoSerial(c, baseProject, fullProjectDirectory, mapWrapper, counter+1)
			}

			// reference to itself
			if nextSuccess == "parallel.json" {
				return DoParallel(c, baseProject, fullProjectDirectory, mapWrapper, counter+1)
			}
		}
		if len(strings.Trim(nextSuccess, " ")) == 0 {
			finalResponseConfigure = cLogicItemTrue.Response
			break
		}

	}

	resultWrapper := response.ParseResponse(mapWrapper, finalResponseConfigure)

	return response.ResponseWriter(resultWrapper, finalResponseConfigure.Transform, c)

}

// worker will called ProcessingRequest. This function is called by parallelRouteHandler function.
func worker(wg *sync.WaitGroup, mapKeyName string, c echo.Context, mapWrapper cmap.ConcurrentMap, requestFromUser model.Wrapper, requestBody []byte, loopIndex int) {
	defer wg.Done()
	_, status, err := ProcessingRequest(mapKeyName, c, requestFromUser, mapWrapper, requestBody, loopIndex)
	if err != nil {
		log.Error("Go Worker - Error Process")
		log.Error(err.Error())
		log.Error("status : ", status)
	}
	//mapWrapper[mapKeyName] = requestFromUser

	mapWrapper.Set(mapKeyName, requestFromUser)
}
