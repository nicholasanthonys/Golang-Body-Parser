package request

import (
	"encoding/json"
	"github.com/jinzhu/copier"
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

		configureItem := configureItem
		loop := DetermineLoop(cc.MapWrapper, configureItem)

		for i := 0; i < loop; i++ {
			alias := configureItem.Alias + "_" + strconv.Itoa(i)
			err := SetRequestToWrapper(alias, cc, &requestFromUser)
			if err != nil {
				log.Error(err.Error())
			}

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
					for _, cLogicItem := range wrapper.Configure.Request.CLogics {
						boolResult, err := service.CLogicsChecker(cLogicItem,
							cc.MapWrapper)
						if err != nil {
							log.Errorf("Error from when checking logic %v", err)
						}
						if boolResult {
							log.Info("CLogic is true for cLogic ", cLogicItem)
							if len(strings.Trim(cLogicItem.NextSuccess, " ")) == 0 {
								log.Info("CLogicItem Response is")

								// if Response is empty
								if reflect.DeepEqual(cLogicItem.Response, model.Command{}) {
									wg.Add(1)
									// process current configure
									go worker(&wg, alias, cc, wrapper, i)
								} else {
									resultWrapper := response.ParseResponse(cc.MapWrapper, cLogicItem.Response, nil, nil)
									response.SetHeaderResponse(resultWrapper.Header, cc)
									return response.ResponseWriter(resultWrapper, cLogicItem.Response.Transform, cc)

								}

							} else {
								wg.Add(1)
								// process next configure
								if wrp, ok := cc.MapWrapper.Get(cLogicItem.NextSuccess); ok {
									newWrapper := wrp.(*model.Wrapper)
									go worker(&wg, cLogicItem.NextSuccess, cc, newWrapper, i)
								} else {
									log.Errorf("cannot load ", cLogicItem.NextSuccess)
								}

							}
						} else {
							log.Info("cLogic is false for clogic ", cLogicItem)
							if len(strings.Trim(cLogicItem.NextFailure, " ")) == 0 {
								log.Info("CLogicItem NExt failure is 0, cLogicItem is ", cLogicItem)
								if !(reflect.DeepEqual(cLogicItem.FailureResponse, model.Command{})) {
									// response
									resultWrapper := response.ParseResponse(cc.MapWrapper, cLogicItem.FailureResponse, nil, nil)
									response.SetHeaderResponse(resultWrapper.Header, cc)
									return response.ResponseWriter(resultWrapper, cLogicItem.FailureResponse.Transform, cc)
								} else {
									continue
								}

							} else {
								wg.Add(1)
								if wrp, ok := cc.MapWrapper.Get(cLogicItem.NextFailure); ok {
									newWrapper := wrp.(*model.Wrapper)
									go worker(&wg, cLogicItem.NextFailure, cc, newWrapper, i)
								} else {
									log.Errorf("cannot load ", cLogicItem.NextFailure)
								}
								//go worker(&wg, cLogicItem.FailureResponse, cc, wrapper, i)
							}
						}

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

	var nextSuccess string
	//finalResponseConfigure := model.Command{}
	for index, cLogicItem := range ParallelProject.CLogics {

		boolResult, err := service.CLogicsChecker(cLogicItem, cc.MapWrapper)
		log.Info("CLogic item is ", cLogicItem, "bool result ", boolResult)
		if err != nil {
			log.Error(err)
			tmpMapResponse := response.ParseResponse(cc.MapWrapper, ParallelProject.FailureResponse, err, nil)
			return response.ResponseWriter(tmpMapResponse, ParallelProject.FailureResponse.Transform, cc)
		}

		if boolResult {
			nextSuccess = cLogicItem.NextSuccess
			if len(strings.Trim(nextSuccess, " ")) > 0 {
				if nextSuccess == "serial.json" {
					return DoSerial(cc, counter+1)
				} else {
					// reference to itself
					return DoParallel(cc, counter+1)
				}

			} else {
				resultWrapper := response.ParseResponse(cc.MapWrapper, cLogicItem.Response, nil, nil)
				response.SetHeaderResponse(resultWrapper.Header, cc)
				return response.ResponseWriter(resultWrapper, cLogicItem.Response.Transform, cc)
			}

		} else {
			if !reflect.DeepEqual(cLogicItem.FailureResponse, model.Command{}) {
				resultWrapper := response.ParseResponse(cc.MapWrapper, cLogicItem.FailureResponse, nil, nil)
				response.SetHeaderResponse(resultWrapper.Header, cc)
				return response.ResponseWriter(resultWrapper, cLogicItem.FailureResponse.Transform, cc)
			} else {
				if index == len(ParallelProject.CLogics)-1 {
					resultWrapper := response.ParseResponse(cc.MapWrapper, ParallelProject.FailureResponse, nil, nil)
					response.SetHeaderResponse(resultWrapper.Header, cc)
					return response.ResponseWriter(resultWrapper, ParallelProject.FailureResponse.Transform, cc)
				}

			}

		}

		//// update next_success
		//nextSuccess = cLogicItem.NextSuccess
		//// update alias
		//if len(strings.Trim(nextSuccess, " ")) > 0 {
		//	if nextSuccess == "serial.json" {
		//		return DoSerial(cc, counter+1)
		//	}
		//
		//	// reference to itself
		//	if nextSuccess == "parallel.json" {
		//		return DoParallel(cc, counter+1)
		//	}
		//}
		//if len(strings.Trim(nextSuccess, " ")) == 0 {
		//	finalResponseConfigure = cLogicItem.Response
		//	break
		//}

	}

	return nil
	//
	//resultWrapper := response.ParseResponse(cc.MapWrapper, finalResponseConfigure, nil, nil)
	//
	//return response.ResponseWriter(resultWrapper, finalResponseConfigure.Transform, cc)

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
