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
	"strings"
)

func DoSerial(c echo.Context, baseProject model.Base, fullProjectDirectory string, mapWrapper cmap.ConcurrentMap, counter int) error {

	if counter == baseProject.MaxCircular {
		if &baseProject.CircularResponse != nil {
			resMap := response.ParseResponse(mapWrapper, baseProject.CircularResponse, nil)
			return response.ResponseWriter(resMap, baseProject.CircularResponse.Transform, c)
		}
		resMap := make(map[string]interface{})
		resMap["message"] = "Circular Request detected"
		return c.JSON(http.StatusLoopDetected, resMap)

	}

	var SerialProject model.Serial
	SerialProject = model.Serial{}
	// Read SerialProject .json
	serialByte := util.ReadJsonFile(fullProjectDirectory + "/" + "serial.json")
	err := json.Unmarshal(serialByte, &SerialProject)

	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In unmarshaling File serial.json. "
		resMap["error"] = err.Error()
		return c.JSON(http.StatusInternalServerError, resMap)
	}

	reqByte, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		resMap := make(map[string]string)
		resMap["message"] = "Problem In Reading Request Body. " + err.Error()
		return c.JSON(http.StatusInternalServerError, resMap)
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

		configByte := util.ReadJsonFile(fullProjectDirectory + "/" + configureItem.FileName)

		//* assign configure byte to configure
		_ = json.Unmarshal(configByte, &configure)
		requestFromUser.Configure = configure

		// store map alias - configure so it is easier to refer
		mapConfigures[configureItem.Alias] = configureItem

		// store map wrapper
		mapWrapper.Set(configureItem.Alias, requestFromUser)

	}

	alias := SerialProject.Configures[0].Alias
	if len(SerialProject.Configures[0].CLogics) == 0 {
		var wrapper model.Wrapper
		if tmp, ok := mapWrapper.Get(alias); ok {
			wrapper = tmp.(model.Wrapper)
		}
		_, _, mapResponse, err := ProcessingRequest(alias, c, wrapper, mapWrapper, reqByte, 0)
		mapWrapper.Set(alias, wrapper)
		if err != nil {
			// next failure
			tmpMapResponse := response.ParseResponse(mapWrapper, mapConfigures[alias].NextFailure, err)
			return response.ResponseWriter(tmpMapResponse, mapConfigures[alias].NextFailure.Transform, c)
		}
		return response.ResponseWriter(mapResponse, wrapper.Configure.Response.Transform, c)

	}

	// assumption :  the first configure to be processed is configures at index 0 from SerialProject.configures
	nextSuccess := SerialProject.Configures[0].CLogics[0].NextSuccess
	finalResponseConfigure := model.Command{}

	for {
		// Processing request

		// Retrieve item from map.
		var wrapper model.Wrapper
		if tmp, ok := mapWrapper.Get(alias); ok {
			wrapper = tmp.(model.Wrapper)
		}
		// Loop only available for parallel request, therefore, set loopIndex to 0
		_, _, _, err := ProcessingRequest(alias, c, wrapper, mapWrapper, reqByte, 0)
		if err != nil {
			// next failure
			tmpMapResponse := response.ParseResponse(mapWrapper, mapConfigures[alias].NextFailure, err)

			return response.ResponseWriter(tmpMapResponse, mapConfigures[alias].NextFailure.Transform, c)
		}

		mapWrapper.Set(alias, wrapper)
		cLogicItemTrue, err := service.CLogicsChecker(mapConfigures[alias].CLogics, mapWrapper)

		if err != nil || cLogicItemTrue == nil {
			log.Error(err)
			tmpMapResponse := response.ParseResponse(mapWrapper, mapConfigures[alias].NextFailure, err)

			return response.ResponseWriter(tmpMapResponse, wrapper.Configure.Response.Transform, c)
		}

		// update next_success
		nextSuccess = cLogicItemTrue.NextSuccess
		// update alias
		if len(strings.Trim(nextSuccess, " ")) > 0 {

			// reference to parallel request
			if nextSuccess == "parallel.json" {
				return DoParallel(c, baseProject, fullProjectDirectory, mapWrapper, counter+1)
			}

			// reference to itself
			if nextSuccess == "serial.json" {
				return DoSerial(c, baseProject, fullProjectDirectory, mapWrapper, counter+1)
			}
			alias = nextSuccess
		}
		if len(strings.Trim(nextSuccess, " ")) == 0 {
			finalResponseConfigure = cLogicItemTrue.Response
			break
		}

	}

	//var wrapper model.Wrapper
	//if tmp, ok := mapWrapper.Get(alias); ok {
	//	wrapper = tmp.(model.Wrapper)
	//}

	tmpMapResponse := response.ParseResponse(mapWrapper, finalResponseConfigure, err)
	return response.ResponseWriter(tmpMapResponse, finalResponseConfigure.Transform, c)
}
