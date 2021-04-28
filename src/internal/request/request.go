package request

import (
	"errors"
	"github.com/labstack/echo"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	responseEntity "github.com/nicholasanthonys/Golang-Body-Parser/internal/response"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
)

var log = logrus.New()

func init() {
	//* init logger with timestamp
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	log.Level = util.GetLogLevelFromEnv()
}

func ParseRequestBody(c echo.Context, contentType string, reqByte []byte) (map[string]interface{}, int, error) {
	var err error
	var result = make(map[string]interface{})
	switch contentType {
	case "application/json":
		//*transform JSON request user to map request from user
		result, err = service.FromJson(reqByte)
		if err != nil {
			logrus.Warn("error parse request body from Json")
			result["message"] = err.Error()
			return nil, http.StatusInternalServerError, err
		}

	case "application/x-www-form-urlencoded":
		//*transform x www form url encoded request user to map request from user
		result = service.FromFormUrl(c)
	case "application/xml":

		//*transform xml request user to map request from user
		result, err = service.FromXmL(reqByte)
		if err != nil {
			logrus.Warn("error service from xml")
			result["message"] = err.Error()
			return nil, http.StatusInternalServerError, err
		}

	default:
		logrus.Warn("Content type not supported")
		return nil, http.StatusBadRequest, errors.New("Content Type Not Supported")
	}
	return result, http.StatusOK, nil
}

// ProcessingRequest is the core function to process every configure. doCommand for transformation, send and receive request happen here.
func ProcessingRequest(aliasName string, c echo.Context, wrapper model.Wrapper, mapWrapper cmap.ConcurrentMap, reqByte []byte, loopIndex int) (*model.Wrapper, int, map[string]interface{}, error) {
	//*check the content type user request
	var contentType string
	var err error
	var status int

	if c.Request().Header["Content-Type"] != nil {
		contentType = c.Request().Header["Content-Type"][0]
	} else {
		contentType = "application/json"
	}

	//*convert request to map string interface based on the content type
	var tmpRequestBody map[string]interface{}
	tmpRequestBody, status, err = ParseRequestBody(c, contentType, reqByte)

	if err != nil {
		return nil, status, nil, err
	}

	//*set header value
	tmpRequestHeader := make(map[string]interface{})
	for key := range c.Request().Header {

		tmpRequestHeader[key] = c.Request().Header.Get(key)
	}

	//*set query value
	tmpRequestQuery := make(map[string]interface{})
	for key := range c.QueryParams() {
		tmpRequestQuery[key] = c.QueryParam(key)
	}

	//*set param value
	tmpRequestParam := make(map[string]interface{})
	for _, value := range c.ParamNames() {
		tmpRequestParam[value] = c.Param(value)
	}

	// write
	wrapper.Request.Set("param", tmpRequestParam)
	wrapper.Request.Set("header", tmpRequestHeader)
	wrapper.Request.Set("body", tmpRequestBody)
	wrapper.Request.Set("query", tmpRequestQuery)

	//* In case user want to log before modify/changing request
	if len(wrapper.Configure.Request.LogBeforeModify) > 0 {
		logValue := make(map[string]interface{}) // value to be logged
		for key, val := range wrapper.Configure.Request.LogBeforeModify {
			logValue[key] = service.RetrieveValue(val, wrapper.Request, loopIndex)
		}
		//logValue = service.RetrieveValue(wrapper.Configure.Request.LogBeforeModify, wrapper.Request, loopIndex)
		util.DoLoggingJson(logValue, "before", aliasName, true)
	}

	//*assign first before do any add,modification,delete in case value want reference each other
	//mapWrapper[aliasName] = *wrapper
	mapWrapper.Set(aliasName, wrapper)

	//* Do the Map Modification
	//var mutex = &sync.Mutex{}
	//mutex.Lock()
	tmpMapRequest := service.DoAddModifyDelete(wrapper.Configure.Request, &wrapper.Request, &mapWrapper, loopIndex)
	// mutex.Unlock()
	//write
	wrapper.Request.Set("header", tmpMapRequest["header"])
	wrapper.Request.Set("body", tmpMapRequest["body"])
	wrapper.Request.Set("query", tmpMapRequest["query"])

	mapWrapper.Set(aliasName, wrapper)

	//*get the destinationPath value before sending request
	wrapper.Configure.Request.DestinationPath = service.ModifyPath(wrapper.Configure.Request.DestinationPath, "--", &mapWrapper, loopIndex)

	//* In case user want to log after modify/changing request
	if len(wrapper.Configure.Request.LogAfterModify) > 0 {
		logValue := make(map[string]interface{}) // value to be logged
		for key, val := range wrapper.Configure.Request.LogAfterModify {
			logValue[key] = service.RetrieveValue(val, wrapper.Request, loopIndex)
		}
		util.DoLoggingJson(logValue, "after", aliasName, true)
	}

	//*send to destination url
	response, err := service.Send(&wrapper)

	if err != nil {
		logrus.Error("Error send : ", err.Error())
		return nil, http.StatusInternalServerError, nil, err
	}
	//*close http
	defer response.Body.Close()

	//*Modify responseByte in Receiver and get  byte from response that has been modified
	var tmpResponse map[string]interface{}
	tmpResponse, err = responseEntity.Receiver(wrapper.Configure, response)

	wrapper.Response.Set("statusCode", tmpResponse["statusCode"])
	wrapper.Response.Set("header", tmpResponse["header"])

	if !reflect.ValueOf(tmpResponse["body"]).IsNil() {
		wrapper.Response.Set("body", tmpResponse["body"])
	}

	if err != nil {
		return nil, http.StatusInternalServerError, nil, err
	}

	//* In case user want to log before modify/changing request
	if len(wrapper.Configure.Response.LogBeforeModify) > 0 {
		logValue := make(map[string]interface{}) // value to be logged
		for key, val := range wrapper.Configure.Response.LogBeforeModify {
			logValue[key] = service.RetrieveValue(val, wrapper.Response, loopIndex)
		}
		util.DoLoggingJson(wrapper.Configure.Response.LogBeforeModify, "before", aliasName, false)
	}

	//* Do Command Add, Modify, Deletion for response again
	tmpMapResponseModified := service.DoAddModifyDelete(wrapper.Configure.Response, &wrapper.Response, &mapWrapper, loopIndex)
	if wrapper.Configure.Request.StatusCode > 0 {
		tmpMapResponseModified["statusCode"] = wrapper.Configure.Response.StatusCode
	} else {
		tmpMapResponseModified["statusCode"] = response.StatusCode
	}

	wrapper.Response.Set("header", tmpMapResponseModified["header"])
	wrapper.Response.Set("body", tmpMapResponseModified["body"])

	//* In case user want to log after modify/changing request
	if len(wrapper.Configure.Response.LogAfterModify) > 0 {
		logValue := make(map[string]interface{}) // value to be logged
		for key, val := range wrapper.Configure.Response.LogAfterModify {
			logValue[key] = service.RetrieveValue(val, wrapper.Response, loopIndex)
		}
		util.DoLoggingJson(logValue, "after", aliasName, false)
	}
	return &wrapper, http.StatusOK, tmpMapResponseModified, nil
}
