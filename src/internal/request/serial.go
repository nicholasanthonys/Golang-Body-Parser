package request

import (
	"encoding/json"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/response"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	cmap "github.com/orcaman/concurrent-map"
	"io/ioutil"
	"net/http"
	"strings"
)

func DoSerial(cc *model.CustomContext, counter int) error {

	if counter == cc.BaseProject.MaxCircular {
		if &cc.BaseProject.CircularResponse != nil {
			resMap := response.ParseResponse(cc.MapWrapper, cc.BaseProject.CircularResponse, nil)
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
		resMap["error"] = err.Error()
		return cc.JSON(http.StatusInternalServerError, resMap)
	}

	reqByte, err := ioutil.ReadAll(cc.Request().Body)
	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In Reading Request Body. " + err.Error()
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
		cc.MapWrapper.Set(configureItem.Alias, requestFromUser)

	}

	alias := SerialProject.Configures[0].Alias
	if len(SerialProject.Configures[0].CLogics) == 0 {
		var wrapper model.Wrapper
		if tmp, ok := cc.MapWrapper.Get(alias); ok {
			wrapper = tmp.(model.Wrapper)
		}
		_, _, mapResponse, err := ProcessingRequest(alias, cc, wrapper, reqByte, 0)
		cc.MapWrapper.Set(alias, wrapper)
		if err != nil {
			// next failure
			tmpMapResponse := response.ParseResponse(cc.MapWrapper, mapConfigures[alias].NextFailure, err)
			return response.ResponseWriter(tmpMapResponse, mapConfigures[alias].NextFailure.Transform, cc)
		}
		return response.ResponseWriter(mapResponse, wrapper.Configure.Response.Transform, cc)

	}

	// assumption :  the first configure to be processed is configures at index 0 from SerialProject.configures
	nextSuccess := SerialProject.Configures[0].CLogics[0].NextSuccess
	finalResponseConfigure := model.Command{}

	for {
		// Processing request

		// Retrieve item from map.
		var wrapper model.Wrapper
		if tmp, ok := cc.MapWrapper.Get(alias); ok {
			wrapper = tmp.(model.Wrapper)
		}
		// Loop only available for parallel request, therefore, set loopIndex to 0
		_, _, _, err := ProcessingRequest(alias, cc, wrapper, reqByte, 0)
		if err != nil {
			// next failure
			tmpMapResponse := response.ParseResponse(cc.MapWrapper, mapConfigures[alias].NextFailure, err)

			return response.ResponseWriter(tmpMapResponse, mapConfigures[alias].NextFailure.Transform, cc)
		}

		cc.MapWrapper.Set(alias, wrapper)
		cLogicItemTrue, err := service.CLogicsChecker(mapConfigures[alias].CLogics, cc.MapWrapper)

		if err != nil || cLogicItemTrue == nil {
			log.Error(err)
			tmpMapResponse := response.ParseResponse(cc.MapWrapper, mapConfigures[alias].NextFailure, err)

			return response.ResponseWriter(tmpMapResponse, wrapper.Configure.Response.Transform, cc)
		}

		// update next_success
		nextSuccess = cLogicItemTrue.NextSuccess
		// update alias
		if len(strings.Trim(nextSuccess, " ")) > 0 {

			// reference to parallel request
			if nextSuccess == "parallel.json" {
				return DoParallel(cc, counter+1)
			}

			// reference to itself
			if nextSuccess == "serial.json" {
				return DoSerial(cc, counter+1)
			}
			alias = nextSuccess
		}
		if len(strings.Trim(nextSuccess, " ")) == 0 {
			finalResponseConfigure = cLogicItemTrue.Response
			break
		}

	}

	//var wrapper model.Wrapper
	//if tmp, ok := cc.MapWrapper.Get(alias); ok {
	//	wrapper = tmp.(model.Wrapper)
	//}

	tmpMapResponse := response.ParseResponse(cc.MapWrapper, finalResponseConfigure, err)
	return response.ResponseWriter(tmpMapResponse, finalResponseConfigure.Transform, cc)
}
