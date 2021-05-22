package request

import (
	"bytes"
	"encoding/json"
	"github.com/jinzhu/copier"
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

	// declare a WaitGroup
	var wg sync.WaitGroup

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

		// assign configure byte to configure
		_ = json.Unmarshal(configByte, &configure)
		requestFromUser.Configure = configure

		reqByte, _ := ioutil.ReadAll(cc.Request().Body)
		// set content back
		cc.Request().Body = ioutil.NopCloser(bytes.NewBuffer(reqByte))

		configureItem := configureItem
		loop := DetermineLoop(cc.MapWrapper, configureItem)

		for i := 0; i < loop; i++ {
			alias := configureItem.Alias + "_" + strconv.Itoa(i)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := SetRequestToWrapper(alias, cc, &requestFromUser, reqByte)
				if err != nil {
					log.Error(err.Error())
				}
			}()

		}
	}

	wg.Wait()

	for _, configureItem := range ParallelProject.Configures {
		loop := DetermineLoop(cc.MapWrapper, configureItem)
		for i := 0; i < loop; i++ {
			alias := configureItem.Alias + "_" + strconv.Itoa(i)
			if wrp, ok := cc.MapWrapper.Get(alias); ok {
				wrapper := wrp.(*model.Wrapper)
				if len(wrapper.Configure.Request.CLogics) > 0 {
					cLogicItem, _ := service.CLogicsChecker(wrapper.Configure.Request.CLogics, cc.MapWrapper)
					if cLogicItem != nil {
						wg.Add(1)
						go worker(&wg, alias, cc, wrapper, i)
					}
				} else {
					// no clogics
					wg.Add(1)
					go worker(&wg, alias, cc, wrapper, i)
				}
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

// worker will called ProcessingRequest. This function is called by parallelRouteHandler function.
func worker(wg *sync.WaitGroup, mapKeyName string, cc *model.CustomContext, requestFromUser *model.Wrapper, loopIndex int) {
	defer wg.Done()

	configure := model.Configure{}
	request := make(map[string]interface{})
	res := make(map[string]interface{})

	err := copier.CopyWithOption(&request, requestFromUser.Request.Items(), copier.Option{DeepCopy: true})
	if err != nil {
		return
	}

	err = copier.CopyWithOption(&res, requestFromUser.Response.Items(), copier.Option{DeepCopy: true})
	if err != nil {
		return
	}

	err = copier.CopyWithOption(&configure, requestFromUser.Configure, copier.Option{DeepCopy: true})
	if err != nil {
		return
	}

	tempRequestFromUser := model.Wrapper{
		Configure: model.Configure{},
		Request:   cmap.New(),
		Response:  cmap.New(),
	}

	tempRequestFromUser.Request.Set("param", request["param"])
	tempRequestFromUser.Request.Set("header", request["header"])
	tempRequestFromUser.Request.Set("body", request["body"])
	tempRequestFromUser.Request.Set("query", request["query"])

	tempRequestFromUser.Configure = configure

	_, status, err := ProcessingRequest(mapKeyName, cc, &tempRequestFromUser, loopIndex)
	if err != nil {
		log.Error("Go Worker - Error Process")
		log.Error(err.Error())
		log.Error("status : ", status)
	}

	cc.MapWrapper.Set(mapKeyName, &tempRequestFromUser)
}

func DetermineLoop(mapWrapper *cmap.ConcurrentMap, configure model.ConfigureItem) int {
	loopIn := service.InterfaceDirectModifier(configure.Loop, mapWrapper, "--")
	lt := reflect.TypeOf(loopIn)
	var loop int
	var err error

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
		log.Warn("set loop to 1 ")
		loop = 1
	}

	return loop
}
