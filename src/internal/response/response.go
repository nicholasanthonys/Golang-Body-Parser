package response

import (
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

// setHeaderResponse set custom key-value pair for header, except Content-Length and Content-type
func SetHeaderResponse(header map[string]interface{}, cc *model.CustomContext) *model.CustomContext {
	for key, val := range header {
		if val == nil {
			log.Warn("set header response for key : ", key, " val is : ", val, " is equal to nil")
		} else {
			rt := reflect.TypeOf(val)
			//* only add if interface type is string
			if rt.Kind() == reflect.String {
				if key != "Content-Length" && key != "Content-Type" {
					cc.Response().Header().Set(key, val.(string))
				}

			} else {
				log.Warn(" set header response for key ", key, " val : ", val, " is not a string")
			}
		}

	}

	return cc

}

// parseResponse process response (add,modify,delete) and return map to be sent to the client
func ParseResponse(mapWrapper *cmap.ConcurrentMap, command model.Command, err error, customResponse *model.CustomResponse) model.CustomResponse {

	resultWrapper := model.Wrapper{
		Configure: model.Configure{
			Response: command,
		},
		Response: cmap.New(),
	}

	resultWrapper.Response.Set("statusCode", 0)
	resultWrapper.Response.Set("header", make(map[string]interface{}))
	resultWrapper.Response.Set("body", make(map[string]interface{}))

	//* now we will set the response body based from configurex.json if there is $configure value in configureBased.

	tmpHeader := make(map[string]interface{})
	tmpBody := make(map[string]interface{})

	statusCode := 400
	if customResponse != nil {
		tmpHeader = customResponse.Header
		tmpBody = customResponse.Body
		if customResponse.StatusCode > 0 {
			statusCode = customResponse.StatusCode

		} else {
			log.Warn("status code is not defined, set status code to 400")
			// default
			statusCode = 400

		}
	}

	// if status code is specified in configure, then set status code based on configure
	if command.StatusCode > 0 {
		statusCode = command.StatusCode
	}

	//*header
	tmpHeader = service.AddToWrapper(resultWrapper.Configure.Response.Adds.Header, "--", tmpHeader, mapWrapper, 0)
	//*modify header
	tmpHeader = service.ModifyWrapper(resultWrapper.Configure.Response.Modifies.Header, "--", tmpHeader, mapWrapper, 0)
	//*Deletion Header
	tmpHeader = service.DeletionHeaderOrQuery(resultWrapper.Configure.Response.Deletes.Header, tmpHeader)

	//*add
	tmpBody = service.AddToWrapper(resultWrapper.Configure.Response.Adds.Body, "--", tmpBody, mapWrapper, 0)
	//*modify
	tmpBody = service.ModifyWrapper(resultWrapper.Configure.Response.Modifies.Body, "--", tmpBody, mapWrapper, 0)
	//* delete
	tmpBody = service.DeletionBody(resultWrapper.Configure.Response.Deletes, tmpBody)

	//*In case user want to log final response
	if len(resultWrapper.Configure.Response.LogAfterModify) > 0 {
		logValue := make(map[string]interface{}) // v
		for key, val := range resultWrapper.Configure.Response.LogAfterModify {
			logValue[key] = service.GetFromHalfReferenceValue(val, resultWrapper.Response, 0)
		}
		//logValue := service.GetFromHalfReferenceValue(resultWrapper.Configure.Response.LogAfterModify, resultWrapper.Response, 0)
		util.DoLoggingJson(logValue, "after", "final response", false)
	}

	response := model.CustomResponse{
		StatusCode: statusCode,
		Header:     tmpHeader,
		Body:       tmpBody,
		Error:      err,
	}

	return response
}

//*ResponseWriter is a function that will return response
func ResponseWriter(customResponse model.CustomResponse, transform string, cc *model.CustomContext) error {
	var statusCode int
	statusCode = customResponse.StatusCode

	responseBody := customResponse.Body
	responseHeader := customResponse.Header

	if customResponse.Error != nil {
		responseBody["error"] = customResponse.Error.Error()
	}

	SetHeaderResponse(responseHeader, cc)
	if statusCode == 0 {
		log.Warn("Status Code is not defined, set Status code to 4000")
		statusCode = 400
	}

	switch strings.ToLower(transform) {
	case strings.ToLower("ToJson"):
		return cc.JSON(statusCode, responseBody)
	case strings.ToLower("ToXml"):

		resByte, err := service.ToXml(responseBody)

		if err != nil {
			log.Error(err.Error())
			res := make(map[string]interface{})
			res["message"] = err.Error()
			return cc.XML(500, res)
		}
		return cc.XMLBlob(statusCode, resByte)
	default:
		logrus.Info("type not supported. only support ToJson and ToXml. Your transform : " + strings.ToLower(transform))
		return cc.JSON(404, "Type Not Supported. only support ToJson and ToXml")
	}
}
