package response

import (
	"github.com/labstack/echo"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
	"reflect"
	"strconv"
	"strings"
)

// setHeaderResponse set custom key-value pair for header, except Content-Length and Content-type
func SetHeaderResponse(header map[string]interface{}, c echo.Context) echo.Context {
	for key, val := range header {
		rt := reflect.TypeOf(val)
		//* only add if interface type is string
		if rt.Kind() == reflect.String {
			if key != "Content-Length" && key != "Content-Type" {
				c.Response().Header().Set(key, val.(string))
			}

		}
	}

	return c

}

// parseResponse process response (add,modify,delete) and return map to be sent to the client
func ParseResponse(mapWrapper cmap.ConcurrentMap, command model.Command, err error) map[string]interface{} {

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
	//keyConfigure := util.RemoveCharacters(resultWrapper.Configure.ConfigureBased, "$")
	//if strings.HasPrefix(resultWrapper.Configure.ConfigureBased, "$configure") {
	//
	//	//*check if key exist in the map
	//	if _, ok := mapWrapper[keyConfigure]; ok {
	//		//* get configureX.json from map wrapper
	//		resultWrapper.Response = mapWrapper[keyConfigure].Response
	//	}
	//}
	tmpHeader := make(map[string]interface{})
	tmpBody := make(map[string]interface{})

	//*header
	tmpHeader = service.AddToWrapper(resultWrapper.Configure.Response.Adds.Header, "--", tmpHeader, &mapWrapper, 0)
	//*modify header
	tmpHeader = service.ModifyWrapper(resultWrapper.Configure.Response.Modifies.Header, "--", tmpHeader, &mapWrapper, 0)
	//*Deletion Header
	tmpHeader = service.DeletionHeaderOrQuery(resultWrapper.Configure.Response.Deletes.Header, tmpHeader)

	//*add
	tmpBody = service.AddToWrapper(resultWrapper.Configure.Response.Adds.Body, "--", tmpBody, &mapWrapper, 0)
	//*modify
	tmpBody = service.ModifyWrapper(resultWrapper.Configure.Response.Modifies.Body, "--", tmpBody, &mapWrapper, 0)
	//* delete
	tmpBody = service.DeletionBody(resultWrapper.Configure.Response.Deletes, tmpBody)

	//*In case user want to log final response
	if len(resultWrapper.Configure.Response.LogAfterModify) > 0 {
		logValue := make(map[string]interface{}) // v
		for key, val := range resultWrapper.Configure.Response.LogAfterModify {
			logValue[key] = service.RetrieveValue(val, resultWrapper.Response, 0)
		}
		//logValue := service.RetrieveValue(resultWrapper.Configure.Response.LogAfterModify, resultWrapper.Response, 0)
		util.DoLoggingJson(logValue, "after", "final response", false)
	}

	var statusCode int
	if resultWrapper.Configure.Response.StatusCode == 0 {
		// default
		statusCode = 400
	} else {
		statusCode = resultWrapper.Configure.Response.StatusCode
	}

	statusCodeString := strconv.Itoa(statusCode)

	response := map[string]interface{}{
		"statusCode": statusCodeString,
		"header":     tmpHeader,
		"body":       tmpBody,
		"error":      err,
	}
	return response
}

//*ResponseWriter is a function that will return response
func ResponseWriter(mapResponse map[string]interface{}, transform string, c echo.Context) error {
	var statusCode int
	statusCode, _ = mapResponse["statusCode"].(int)
	responseBody := mapResponse["body"].(map[string]interface{})
	responseHeader := mapResponse["header"].(map[string]interface{})

	if mapResponse["error"] != nil {
		responseBody["error"] = mapResponse["error"].(error).Error()
	}

	c = SetHeaderResponse(responseHeader, c)
	if statusCode == 0 {
		statusCode = 200
	}

	switch strings.ToLower(transform) {
	case strings.ToLower("ToJson"):
		return c.JSON(statusCode, responseBody)
	case strings.ToLower("ToXml"):

		resByte, err := service.ToXml(responseBody)

		if err != nil {
			log.Error(err.Error())
			res := make(map[string]interface{})
			res["message"] = err.Error()
			return c.XML(500, res)
		}
		return c.XMLBlob(statusCode, resByte)
	default:
		logrus.Info("type not supported. only support ToJson and ToXml. Your transform : " + strings.ToLower(transform))
		return c.JSON(404, "Type Not Supported. only support ToJson and ToXml")
	}
}
