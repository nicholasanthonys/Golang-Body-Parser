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

func DoSerial(c echo.Context, fullProjectDirectory string, mapWrapper cmap.ConcurrentMap, counter int) error {
	if counter == 10 {
		resMap := make(map[string]string)
		resMap["message"] = "Circular Serial-Parallel"
		return c.JSON(http.StatusInternalServerError, resMap)
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

		requestFromUser.Response.Set("statusCode", "")
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
		//mapWrapper[configureItem.Alias] = requestFromUser

	}

	// assumption :  the first configure to be processed is configures at index 0 from SerialProject.configures
	nextSuccess := SerialProject.Configures[0].CLogics[0].NextSuccess
	alias := SerialProject.Configures[0].Alias
	finalResponseConfigure := model.Command{}

	for {
		// Processing request

		// Retrieve item from map.
		var wrapper model.Wrapper
		if tmp, ok := mapWrapper.Get(alias); ok {
			wrapper = tmp.(model.Wrapper)
		}
		// Loop only available for parallel request, therefore, set loopIndex to 0
		_, _, err := ProcessingRequest(alias, c, wrapper, mapWrapper, reqByte, 0)
		if err != nil {
			// next failure
			tmpMapResponse := response.ParseResponse(mapWrapper, mapConfigures[alias].NextFailure)
			tmpMapResponse["header"] = response.SetHeaderResponse(tmpMapResponse["header"].(map[string]interface{}), c)

			return util.ResponseWriter(tmpMapResponse, wrapper.Configure.Response.Transform, c)
		}

		mapWrapper.Set(alias, wrapper)
		cLogicItemTrue, err := service.CLogicsChecker(mapConfigures[alias].CLogics, mapWrapper)

		if err != nil || cLogicItemTrue == nil {
			log.Error(err)
			tmpMapResponse := response.ParseResponse(mapWrapper, mapConfigures[alias].NextFailure)
			tmpMapResponse["header"] = response.SetHeaderResponse(tmpMapResponse["header"].(map[string]interface{}), c)

			return util.ResponseWriter(tmpMapResponse, wrapper.Configure.Response.Transform, c)
		}

		// update next_success
		nextSuccess = cLogicItemTrue.NextSuccess
		// update alias
		if len(strings.Trim(nextSuccess, " ")) > 0 {
			if nextSuccess == "parallel.json" {
				return DoParallel(c, fullProjectDirectory, mapWrapper, counter+1)
			}
			alias = nextSuccess
		}
		if len(strings.Trim(nextSuccess, " ")) == 0 {
			finalResponseConfigure = cLogicItemTrue.Response
			break
		}

	}

	var wrapper model.Wrapper
	if tmp, ok := mapWrapper.Get(alias); ok {
		wrapper = tmp.(model.Wrapper)
	}

	tmpMapResponse := response.ParseResponse(mapWrapper, finalResponseConfigure)
	tmpMapResponse["header"] = response.SetHeaderResponse(tmpMapResponse["header"].(map[string]interface{}), c)

	return util.ResponseWriter(tmpMapResponse, wrapper.Configure.Response.Transform, c)
}
