package response

import (
	"github.com/labstack/echo"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
	"reflect"
	"strconv"
)

// setHeaderResponse set custom key-value pair for header, except Content-Length and Content-type
func SetHeaderResponse(header map[string]interface{}, c echo.Context) {
	for key, val := range header {
		rt := reflect.TypeOf(val)
		//* only add if interface type is string
		if rt.Kind() == reflect.String {
			if key != "Content-Length" && key != "Content-Type" {
				c.Response().Header().Set(key, val.(string))
			}

		}
	}

}

// parseResponse process response (add,modify,delete) and return map to be sent to the client
func ParseResponse(mapWrapper map[string]model.Wrapper, command model.Command) model.Wrapper {

	resultWrapper := model.Wrapper{
		Configure: model.Configure{
			Response: command,
		},
		Request: model.Fields{},
		Response: model.Fields{
			Param:  make(map[string]interface{}),
			Header: make(map[string]interface{}),
			Body:   make(map[string]interface{}),
			Query:  make(map[string]interface{}),
		},
	}

	//* now we will set the response body based from configurex.json if there is $configure value in configureBased.
	//keyConfigure := util.RemoveCharacters(resultWrapper.Configure.ConfigureBased, "$")
	//if strings.HasPrefix(resultWrapper.Configure.ConfigureBased, "$configure") {
	//
	//	//*check if key exist in the map
	//	if _, ok := mapWrapper[keyConfigure]; ok {
	//		//* get configureX.json from map wrapper
	//		resultWrapper.Response = mapWrapper[keyConfigure].Response
	//	}
	//}

	logrus.Info("result wrapper configur response transform")
	logrus.Info(resultWrapper.Configure.Response.Transform)
	//*header
	resultWrapper.Response.Header = service.AddToWrapper(resultWrapper.Configure.Response.Adds.Header, "--", resultWrapper.Response.Header, mapWrapper, 0)
	//*modify header
	resultWrapper.Response.Header = service.ModifyWrapper(resultWrapper.Configure.Response.Modifies.Header, "--", resultWrapper.Response.Header, mapWrapper, 0)
	//*Deletion Header
	resultWrapper.Response.Header = service.DeletionHeaderOrQuery(resultWrapper.Configure.Response.Deletes.Header, resultWrapper.Response.Header)

	//*add
	resultWrapper.Response.Body = service.AddToWrapper(resultWrapper.Configure.Response.Adds.Body, "--", resultWrapper.Response.Body, mapWrapper, 0)
	//*modify
	resultWrapper.Response.Body = service.ModifyWrapper(resultWrapper.Configure.Response.Modifies.Body, "--", resultWrapper.Response.Body, mapWrapper, 0)
	//* delete
	resultWrapper.Response.Body = service.DeletionBody(resultWrapper.Configure.Response.Deletes, resultWrapper.Response.Body)

	//*In case user want to log final response
	if len(resultWrapper.Configure.Response.LogAfterModify) > 0 {
		logValue := service.RetrieveValue(resultWrapper.Configure.Response.LogAfterModify, resultWrapper.Response, 0)
		util.DoLogging(logValue, "after", "final response", false)
	}

	var statusCode int
	if resultWrapper.Configure.Response.StatusCode == 0 {
		// default
		statusCode = 400
	} else {
		statusCode = resultWrapper.Configure.Response.StatusCode
	}

	resultWrapper.Response.StatusCode = strconv.Itoa(statusCode)

	return resultWrapper
}
