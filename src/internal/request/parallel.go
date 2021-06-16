package request

import (
	"encoding/json"
	"errors"
	"github.com/jinzhu/copier"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	CustomPrometheus "github.com/nicholasanthonys/Golang-Body-Parser/internal/prometheus"
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
			return response.ConstructResponseFromWrapper(cc, cc.BaseProject.CircularResponse, nil, nil)
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
		CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"ERR_UNMARSHAL_PARALLEL_JSON"].Inc()
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
			var alias string
			if loop > 1 {
				alias = configureItem.Alias + "_" + strconv.Itoa(i)
			} else {
				alias = configureItem.Alias
			}
			err := SetRequestToWrapper(alias, cc, &requestFromUser)
			if err != nil {
				log.Errorf("error %s", err.Error())
				CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"ERR_SET_REQUEST_TO_WRAPPER"].Inc()

			}

		}
	}

	wg.Wait()

	for _, configureItem := range ParallelProject.Configures {
		loop := DetermineLoop(cc.MapWrapper, configureItem)
		for i := 0; i < loop; i++ {

			var alias string
			if loop > 1 {
				alias = configureItem.Alias + "_" + strconv.Itoa(i)
			} else {
				alias = configureItem.Alias
			}

			if wrp, ok := cc.MapWrapper.Get(alias); ok {
				wrapper := wrp.(*model.Wrapper)

				if len(wrapper.Configure.Request.CLogics) > 0 {
				CLogics:
					for _, cLogicItem := range wrapper.Configure.Request.CLogics {
						boolResult, err := service.CLogicsChecker(cLogicItem,
							cc.MapWrapper, cc.DefinedRoute.ProjectDirectory)
						if err != nil {
							log.Errorf("Error from when checking logic %v", err)
							CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"ERR_CHECK_CONFIGURE_LOGIC"].Inc()

						}
						if boolResult {
							if len(strings.Trim(cLogicItem.NextSuccess, " ")) == 0 {

								// if Response is empty
								if reflect.DeepEqual(cLogicItem.Response, model.Command{}) {
									wg.Add(1)
									// process current configure
									go worker(&wg, alias, cc, wrapper, i)
								} else {
									return response.ConstructResponseFromWrapper(cc, cLogicItem.Response, nil, nil)
								}

							} else {
								// process next configure
								if wrp, ok := cc.MapWrapper.Get(cLogicItem.NextSuccess); ok {
									wg.Add(1)
									newWrapper := wrp.(*model.Wrapper)
									go worker(&wg, cLogicItem.NextSuccess, cc, newWrapper, i)
								} else {
									log.Errorf("cannot get wrapper : %s", cLogicItem.NextSuccess)
									CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"ERR_GET_WRAPPER"].Inc()

								}

							}
						} else {
							if len(strings.Trim(cLogicItem.NextFailure, " ")) == 0 {
								if !(reflect.DeepEqual(cLogicItem.FailureResponse, model.Command{})) {
									// response
									return response.ConstructResponseFromWrapper(cc, cLogicItem.FailureResponse, nil, nil)

								} else {
									continue CLogics
								}

							} else {
								wg.Add(1)
								if wrp, ok := cc.MapWrapper.Get(cLogicItem.NextFailure); ok {
									newWrapper := wrp.(*model.Wrapper)
									go worker(&wg, cLogicItem.NextFailure, cc, newWrapper, i)
								} else {
									log.Errorf("cannot get wrapper %s", cLogicItem.NextFailure)
									CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"ERR_GET_WRAPPER"].Inc()
									return response.ConstructResponseFromWrapper(cc, cLogicItem.FailureResponse, nil, nil)
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
	for index, cLogicItem := range ParallelProject.CLogics {

		boolResult, err := service.CLogicsChecker(cLogicItem, cc.MapWrapper, cc.DefinedRoute.ProjectDirectory)
		if err != nil {
			log.Error(err)
			CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"ERR_CHECK_CONFIGURE_LOGIC"].Inc()
			return response.ConstructResponseFromWrapper(cc, ParallelProject.FailureResponse, err, nil)

		}

		if boolResult {
			nextSuccess = cLogicItem.NextSuccess
			if len(strings.Trim(nextSuccess, " ")) > 0 {
				if nextSuccess == "serial.json" {
					return DoSerial(cc, counter+1)
				} else if nextSuccess == "parallel.json" {
					// reference to itself
					return DoParallel(cc, counter+1)
				}

				CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"PARALLEL_ERR_INVALID_REFER"].Inc()
				return response.ConstructResponseFromWrapper(cc, cLogicItem.Response,
					errors.New("Parallel can only refer to parallel/serial.json"), nil)

			} else {
				return response.ConstructResponseFromWrapper(cc, cLogicItem.Response, nil, nil)

			}

		} else {
			if !reflect.DeepEqual(cLogicItem.FailureResponse, model.Command{}) {
				return response.ConstructResponseFromWrapper(cc, cLogicItem.FailureResponse, nil, nil)

			} else {
				if index == len(ParallelProject.CLogics)-1 {
					return response.ConstructResponseFromWrapper(cc, ParallelProject.FailureResponse, nil, nil)
				}
			}

		}

	}

	log.Warn("No Response Specified, returning : ", http.StatusBadRequest)
	return cc.JSON(400, map[string]interface{}{
		"message": "No parallel logic to determine response to be returned",
	})
}

// worker will called ProcessingRequest. This function is called by parallelRouteHandler function.
func worker(wg *sync.WaitGroup, mapKeyName string, cc *model.CustomContext, requestFromUser *model.Wrapper, loopIndex int) {
	defer wg.Done()

	configure := model.Configure{}
	request := make(map[string]interface{})
	res := make(map[string]interface{})

	err := copier.CopyWithOption(&request, requestFromUser.Request.Items(), copier.Option{DeepCopy: true})
	if err != nil {
		log.Errorf("error : %s", err.Error())
		CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"PARALLEL_ERR_COPY_REQUEST"].Inc()
		return
	}

	err = copier.CopyWithOption(&res, requestFromUser.Response.Items(), copier.Option{DeepCopy: true})
	if err != nil {
		log.Errorf("error : %s", err.Error())
		CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"PARALLEL_ERR_COPY_RESPONSE"].Inc()
		return
	}

	err = copier.CopyWithOption(&configure, requestFromUser.Configure, copier.Option{DeepCopy: true})
	if err != nil {
		log.Errorf("error : %s", err.Error())
		CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"PARALLEL_ERR_COPY_CONFIGURE"].Inc()
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
			log.Errorf("error:  %s ", err.Error())
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
