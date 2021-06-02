package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
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
			resMap := response.ParseResponse(cc.MapWrapper, cc.BaseProject.CircularResponse, nil, nil)
			return response.ResponseWriter(resMap, cc.BaseProject.CircularResponse.Transform, cc)
		}
		resMap := make(map[string]interface{})
		resMap["message"] = "Circular Request detected"
		return cc.JSON(http.StatusLoopDetected, resMap)

	}

	var SerialProject model.Serial
	SerialProject = model.Serial{}
	// Read SerialProject .json
	serialByte := util.ReadJsonFile(cc.FullProjectDirectory + "/" + "serial.json")
	err := json.Unmarshal(serialByte, &SerialProject)

	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In unmarshaling File serial.json. "
		return cc.JSON(http.StatusInternalServerError, resMap)
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

	//finalResponseConfigure := model.Command{}
	var finalCustomResponse *model.CustomResponse

	// Processing request
	for {
		if tmp, ok := cc.MapWrapper.Get(alias); ok {
			wrapper = tmp.(*model.Wrapper)
		}

		if wrapper == nil {
			log.Errorf("Wrapper is nil for alias : %v ", alias)
			err := errors.New(fmt.Sprintf("wrapper is nil for alias %v", alias))
			return cc.JSON(http.StatusBadRequest, err)
		}

		err = SetRequestToWrapper(alias, cc, wrapper)
		if err != nil {
			return err
		}

		if len(wrapper.Configure.Request.CLogics) > 0 {
			for _, cLogicItem := range wrapper.Configure.Request.CLogics {
				boolResult, err := service.CLogicsChecker(cLogicItem, cc.MapWrapper)
				if err != nil {
					log.Errorf("Error while check logic for cLogic %v : %v", cLogicItem, err)
					return cc.JSON(http.StatusBadRequest, err)
				}

				if boolResult {
					if len(strings.Trim(cLogicItem.NextSuccess, " ")) > 0 {
						alias = nextSuccess
						continue
					} else {
						// if response is not empty
						if !reflect.DeepEqual(cLogicItem.Response, model.Command{}) {
							// boolean result is true and response is specified
							tmpMapResponse := response.ParseResponse(cc.MapWrapper, cLogicItem.Response,
								err, nil)
							return response.ResponseWriter(tmpMapResponse, cLogicItem.Response.Transform, cc)
						} else {
							_, customResponse, err := ProcessingRequest(alias, cc, wrapper, 0)
							finalCustomResponse = customResponse
							if err != nil {
								// next failure
								tmpMapResponse := response.ParseResponse(cc.MapWrapper, mapConfigures[alias].FailureResponse, err, customResponse)
								return response.ResponseWriter(tmpMapResponse, mapConfigures[alias].FailureResponse.Transform, cc)
							}
						}
					}

				} else {
					if len(strings.Trim(cLogicItem.NextFailure, " ")) > 0 {
						// update alias
						alias = nextSuccess
						continue // skip loop and update alias
					} else {
						if !reflect.DeepEqual(cLogicItem.FailureResponse, model.Command{}) {
							// next failure
							tmpMapResponse := response.ParseResponse(cc.MapWrapper, cLogicItem.FailureResponse,
								err, nil)
							return response.ResponseWriter(tmpMapResponse, cLogicItem.FailureResponse.Transform, cc)
						}
					}
				}
			}

		} else {
			_, customResponse, err := ProcessingRequest(alias, cc, wrapper, 0)
			finalCustomResponse = customResponse
			if err != nil {
				log.Errorf("Error after processing request for alias %v:  %v", alias, err)
				// next failure
				tmpMapResponse := response.ParseResponse(cc.MapWrapper, mapConfigures[alias].FailureResponse, err, customResponse)
				return response.ResponseWriter(tmpMapResponse, mapConfigures[alias].FailureResponse.Transform, cc)
			}
		}

		cc.MapWrapper.Set(alias, wrapper)

		if len(mapConfigures[alias].CLogics) == 0 {
			tmpMapResponse := response.ParseResponse(cc.MapWrapper, wrapper.Configure.Response, nil, finalCustomResponse)
			return response.ResponseWriter(tmpMapResponse, wrapper.Configure.Response.Transform, cc)
		}

		for i := 0; i < len(mapConfigures[alias].CLogics); {
			cLogicItem := mapConfigures[alias].CLogics[i]
			boolResult, err := service.CLogicsChecker(cLogicItem, cc.MapWrapper)
			if err != nil {
				log.Error(err)
				tmpMapResponse := response.ParseResponse(cc.MapWrapper, mapConfigures[alias].FailureResponse, err, nil)
				return response.ResponseWriter(tmpMapResponse, wrapper.Configure.Response.Transform, cc)
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
					tmpMapResponse := response.ParseResponse(cc.MapWrapper, cLogicItem.Response, err, finalCustomResponse)
					return response.ResponseWriter(tmpMapResponse, cLogicItem.Response.Transform, cc)
				}
			} else {
				if len(strings.Trim(cLogicItem.NextFailure, " ")) > 0 {
					// update alias
					alias = cLogicItem.NextFailure
					break
				} else {
					if !reflect.DeepEqual(cLogicItem.FailureResponse, model.Command{}) {
						resultWrapper := response.ParseResponse(cc.MapWrapper, cLogicItem.FailureResponse, nil, nil)
						response.SetHeaderResponse(resultWrapper.Header, cc)
						return response.ResponseWriter(resultWrapper, cLogicItem.FailureResponse.Transform, cc)
					} else {
						if i == len(mapConfigures[alias].CLogics)-1 {
							resultWrapper := response.ParseResponse(cc.MapWrapper,
								mapConfigures[alias].FailureResponse, nil, nil)
							response.SetHeaderResponse(resultWrapper.Header, cc)
							return response.ResponseWriter(resultWrapper, mapConfigures[alias].FailureResponse.Transform, cc)
						}
						i++
					}

				}
			}
		}

		// update next_success
		//nextSuccess = cLogicItem.NextSuccess
		// update alias
		//if len(strings.Trim(nextSuccess, " ")) > 0 {
		//
		//	// reference to parallel request
		//	if nextSuccess == "parallel.json" {
		//		return DoParallel(cc, counter+1)
		//	}
		//
		//	// reference to itself
		//	if nextSuccess == "serial.json" {
		//		return DoSerial(cc, counter+1)
		//	}
		//	alias = nextSuccess
		//}
		//if len(strings.Trim(nextSuccess, " ")) == 0 {
		//	finalResponseConfigure = cLogicItem.Response
		//	break
		//}

	}

	//tmpMapResponse := response.ParseResponse(cc.MapWrapper, finalResponseConfigure, err, finalCustomResponse)
	//return response.ResponseWriter(tmpMapResponse, finalResponseConfigure.Transform, cc)
}
