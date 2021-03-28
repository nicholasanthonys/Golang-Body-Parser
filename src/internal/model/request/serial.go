package request

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model/response"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

func DoSerial(c echo.Context, fullProjectDirectory string, mapWrapper map[string]model.Wrapper) error {
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
			Request: model.Fields{
				Param:  make(map[string]interface{}),
				Header: make(map[string]interface{}),
				Body:   make(map[string]interface{}),
				Query:  make(map[string]interface{}),
			},
			Response: model.Fields{
				Param:  make(map[string]interface{}),
				Header: make(map[string]interface{}),
				Body:   make(map[string]interface{}),
				Query:  make(map[string]interface{}),
			},
		}
		configByte := util.ReadJsonFile(fullProjectDirectory + "/" + configureItem.FileName)

		//* assign configure byte to configure
		_ = json.Unmarshal(configByte, &configure)
		requestFromUser.Configure = configure

		// store map alias - configure so it is easier to refer
		mapConfigures[configureItem.Alias] = configureItem

		// store map wrapper
		mapWrapper[configureItem.Alias] = requestFromUser

	}

	// assumption :  the first configure to be processed is configures at index 0 from SerialProject.configures
	nextSuccess := SerialProject.Configures[0].CLogics[0].NextSuccess
	alias := SerialProject.Configures[0].Alias
	finalResponseConfigure := model.Command{}

	for {
		// Processing request
		requestFromUser := mapWrapper[alias]
		// Loop only available for parallel request, therefore, set loopIndex to 0
		_, _, err := ProcessingRequest(alias, c, &requestFromUser, mapWrapper, reqByte, 0)
		if err != nil {
			// next failure
			resultWrapper := response.ParseResponse(mapWrapper, mapConfigures[alias].NextFailure)
			response.SetHeaderResponse(resultWrapper.Response.Header, c)
			return util.ResponseWriter(resultWrapper, c)
			//return util.ErrorWriter(c, requestFromUser.Configure, err, status)
		}

		mapWrapper[alias] = requestFromUser
		cLogicItemTrue, err := service.CLogicsChecker(mapConfigures[alias].CLogics, mapWrapper)

		if err != nil {
			logrus.Error(err)
			resultWrapper := response.ParseResponse(mapWrapper, mapConfigures[alias].NextFailure)
			response.SetHeaderResponse(resultWrapper.Response.Header, c)
			return util.ResponseWriter(resultWrapper, c)
		}

		if cLogicItemTrue == nil {
			resultWrapper := response.ParseResponse(mapWrapper, mapConfigures[alias].NextFailure)
			response.SetHeaderResponse(resultWrapper.Response.Header, c)
			return util.ResponseWriter(resultWrapper, c)
		}

		// update next_success
		nextSuccess = cLogicItemTrue.NextSuccess
		// update alias
		if len(strings.Trim(nextSuccess, " ")) > 0 {
			if nextSuccess == "parallel.json" {
				return DoParallel(c, fullProjectDirectory, mapWrapper)
			}
			alias = nextSuccess
		}
		if len(strings.Trim(nextSuccess, " ")) == 0 {
			finalResponseConfigure = cLogicItemTrue.Response
			break
		}

	}

	resultWrapper := response.ParseResponse(mapWrapper, finalResponseConfigure)
	response.SetHeaderResponse(resultWrapper.Response.Header, c)
	return util.ResponseWriter(resultWrapper, c)
}
