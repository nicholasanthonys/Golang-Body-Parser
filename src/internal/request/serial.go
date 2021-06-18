package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	CustomPrometheus "github.com/nicholasanthonys/Golang-Body-Parser/internal/prometheus"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/response"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	cmap "github.com/orcaman/concurrent-map"
	"net/http"
	"reflect"
	"strings"
)

func DoSerial(cc *model.CustomContext, counter int) error {

	if counter == cc.BaseProject.MaxCircular {
		if &cc.BaseProject.CircularResponse != nil {
			return response.ConstructResponseFromWrapper(cc, cc.BaseProject.CircularResponse, nil, nil)
		}
		resMap := make(map[string]interface{})
		resMap["message"] = "Circular Request detected"
		return cc.JSON(http.StatusLoopDetected, resMap)

	}

	SerialProject := model.Serial{}
	// Read SerialProject .json
	serialByte := util.ReadJsonFile(cc.FullProjectDirectory + "/" + "serial.json")
	err := json.Unmarshal(serialByte, &SerialProject)

	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In unmarshaling File serial.json. "
		CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"ERR_UNMARSHAL_SERIAL_JSON"].Inc()
		return cc.JSON(http.StatusInternalServerError, resMap)
	}

	// set request to map wrapper
	err = SetRequestToWrapper("$configure_request", cc, &model.Wrapper{
		Configure: model.Configure{},
		Request:   cmap.New(),
		Response:  cmap.New(),
	})
	if err != nil {
		log.Errorf(" Error : %s", err.Error())
		CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"ERR_SET_REQUEST_TO_WRAPPER"].Inc()
		return err
	}

	//*Read file ConfigureBased
	var mapConfigures = make(map[string]model.ConfigureItem)

	for _, configureItem := range SerialProject.Configures {
		//read actual configure based on configureItem.file_name
		// Initialization configure object
		var configure = model.Configure{}
		requestFromUser := model.Wrapper{
			Configure: model.Configure{},
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

		// store map alias - configure so it is easier to refer
		mapConfigures[configureItem.Alias] = configureItem

		// store map wrapper
		cc.MapWrapper.Set(configureItem.Alias, &requestFromUser)

	}

	alias := SerialProject.Configures[0].Alias
	var wrapper *model.Wrapper

	// assumption :  the first configure to be processed is configures at index 0 from SerialProject.configures
	var nextSuccess string
	if len(SerialProject.Configures[0].CLogics) > 0 {
		nextSuccess = SerialProject.Configures[0].CLogics[0].NextSuccess
	}

	var finalCustomResponse *model.CustomResponse
	// Processing request

ConfigureFile:
	for {
		if tmp, ok := cc.MapWrapper.Get(alias); ok {
			wrapper = tmp.(*model.Wrapper)
		} else {
			log.Errorf("Wrapper is nil for alias : %v ", alias)
			err := errors.New(fmt.Sprintf("wrapper is nil for alias %v", alias))
			CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"ERR_GET_WRAPPER"].Inc()
			return cc.JSON(http.StatusBadRequest, err)
		}

		if len(wrapper.Configure.Request.CLogics) > 0 {
			for index, cLogicItem := range wrapper.Configure.Request.CLogics {
				boolResult, err := service.CLogicsChecker(cLogicItem, cc.MapWrapper, cc.DefinedRoute.ProjectDirectory)
				if err != nil {
					log.Errorf("Error while check logic for cLogic %v : %v", cLogicItem, err)
					CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"ERR_CHECK_CONFIGURE_LOGIC"].Inc()

					return cc.JSON(http.StatusBadRequest, err)
				}

				if boolResult {
					if len(strings.Trim(cLogicItem.NextSuccess, " ")) > 0 {
						nextSuccess = cLogicItem.NextSuccess
						alias = nextSuccess
						continue ConfigureFile
					} else {
						// if response is not empty
						if !reflect.DeepEqual(cLogicItem.Response, model.Command{}) {
							return response.ConstructResponseFromWrapper(cc, cLogicItem.Response,
								err, nil)
						} else {
							_, customResponse, err := ProcessingRequest(alias, cc, wrapper, 0)
							finalCustomResponse = customResponse
							if err != nil {
								log.Errorf("Error :  %s", err.Error())
								return response.ConstructResponseFromWrapper(cc, mapConfigures[alias].FailureResponse, err, customResponse)
							}
						}
					}

				} else {
					if len(strings.Trim(cLogicItem.NextFailure, " ")) > 0 {
						// update alias
						alias = cLogicItem.NextFailure
						continue ConfigureFile // skip loop and update alias
					} else {
						if !reflect.DeepEqual(cLogicItem.FailureResponse, model.Command{}) {
							return response.ConstructResponseFromWrapper(cc, cLogicItem.FailureResponse,
								err, nil)
						} else {
							if index == len(wrapper.Configure.Request.CLogics)-1 {
								return response.ConstructResponseFromWrapper(cc,
									mapConfigures[alias].FailureResponse, err, &model.CustomResponse{
										StatusCode: 0,
										Header:     nil,
										Body: map[string]interface{}{
											"message": "All logic failed in request " + alias,
										},
										Error: nil,
									})
							}

						}
					}
				}
			}

		} else {
			_, customResponse, err := ProcessingRequest(alias, cc, wrapper, 0)
			finalCustomResponse = customResponse
			if err != nil {
				log.Errorf("Error after processing request for alias %v:  %v", alias, err)
				return response.ConstructResponseFromWrapper(cc, mapConfigures[alias].FailureResponse,
					err, nil)
			}
		}

		cc.MapWrapper.Set(alias, wrapper)

		if len(mapConfigures[alias].CLogics) == 0 {
			return response.ConstructResponseFromWrapper(cc, wrapper.Configure.Response, nil, finalCustomResponse)
		}

		for i := 0; i < len(mapConfigures[alias].CLogics); {
			cLogicItem := mapConfigures[alias].CLogics[i]
			boolResult, err := service.CLogicsChecker(cLogicItem, cc.MapWrapper, cc.DefinedRoute.ProjectDirectory)
			if err != nil {
				log.Errorf("error %s", err.Error())
				CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(cc.DefinedRoute.ProjectDirectory)+"ERR_CHECK_CONFIGURE_LOGIC"].Inc()

				return response.ConstructResponseFromWrapper(cc, mapConfigures[alias].FailureResponse, err, nil)
			}

			// if cLogicItem is not nil and error is nil
			if boolResult {
				nextSuccess = cLogicItem.NextSuccess
				if len(strings.Trim(nextSuccess, " ")) > 0 {

					// reference to parallel request
					if nextSuccess == "parallel.json" {
						return DoParallel(cc, counter+1)
					}

					// reference to itself
					if nextSuccess == "serial.json" {
						return DoSerial(cc, counter+1)
					}

					// update alias
					alias = nextSuccess
					break
				} else {
					return response.ConstructResponseFromWrapper(cc, cLogicItem.Response, err, nil)
				}
			} else {
				if len(strings.Trim(cLogicItem.NextFailure, " ")) > 0 {
					// update alias
					alias = cLogicItem.NextFailure
					break
				} else {
					if !reflect.DeepEqual(cLogicItem.FailureResponse, model.Command{}) {
						return response.ConstructResponseFromWrapper(cc, cLogicItem.FailureResponse, nil, nil)
					} else {
						if i == len(mapConfigures[alias].CLogics)-1 {
							return response.ConstructResponseFromWrapper(cc,
								mapConfigures[alias].FailureResponse, nil, nil)
						}
						i++
					}
				}
			}
		}

	}

}
