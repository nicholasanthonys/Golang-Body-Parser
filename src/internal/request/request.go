package request

import (
	"bytes"
	"errors"
	"github.com/jinzhu/copier"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	CustomPrometheus "github.com/nicholasanthonys/Golang-Body-Parser/internal/prometheus"
	responseEntity "github.com/nicholasanthonys/Golang-Body-Parser/internal/response"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
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

func ParseRequestBody(cc *model.CustomContext, contentType string) (map[string]interface{}, int, error) {
	var err error
	var result = make(map[string]interface{})

	// init variable
	tempCC := cc
	err = copier.Copy(&tempCC, &cc)
	if err != nil {
		log.Error("error copy context in parseRequestBody. error : ", err.Error())
		return nil, 0, err
	}

	reqByte, _ := ioutil.ReadAll(tempCC.Request().Body)
	tempCC.Request().Body = ioutil.NopCloser(bytes.NewBuffer(reqByte))

	switch contentType {
	case "application/json":
		//*transform JSON request user to map request from user
		result, err = service.FromJson(reqByte)
		if err != nil {
			logrus.Warn("error parse request body from Json")
			result["message"] = err.Error()
			CustomPrometheus.PromMapCounter[CustomPrometheus.Prefix+cc.DefinedRoute.ProjectDirectory+"ERR_PARSE_REQUEST_JSON"].Inc()

			return nil, http.StatusInternalServerError, err
		}

	case "application/x-www-form-urlencoded":
		//*transform x www form url encoded request user to map request from user
		result = service.FromFormUrl(cc)
	case "application/xml":

		//*transform xml request user to map request from user
		result, err = service.FromXmL(reqByte)
		if err != nil {
			CustomPrometheus.PromMapCounter[CustomPrometheus.Prefix+cc.DefinedRoute.ProjectDirectory+"ERR_PARSE_REQUEST_XML"].Inc()

			logrus.Warn("error service from xml")
			result["message"] = err.Error()
			return nil, http.StatusInternalServerError, err
		}

	default:
		logrus.Warn("Content type not supported")
		CustomPrometheus.PromMapCounter[CustomPrometheus.Prefix+cc.DefinedRoute.ProjectDirectory+"UNKNOWN_REQUEST_CONTENT_TYPE"].Inc()

		return nil, http.StatusBadRequest, errors.New("Content Type Not Supported")
	}
	return result, http.StatusOK, nil
}

func SetRequestToWrapper(aliasName string, cc *model.CustomContext, wrapper *model.Wrapper) error {

	var contentType string
	var err error

	if cc.Request().Header["Content-Type"] != nil {
		contentType = cc.Request().Header["Content-Type"][0]
	} else {
		contentType = "application/json"
	}

	//*convert request to map string interface based on the content type
	var tmpRequestBody map[string]interface{}

	tmpRequestBody, _, err = ParseRequestBody(cc, contentType)

	if err != nil {
		return err
	}

	//*set header value
	tmpRequestHeader := make(map[string]interface{})
	for key := range cc.Request().Header {
		tmpRequestHeader[key] = cc.Request().Header.Get(key)
	}

	//*set query value
	tmpRequestQuery := make(map[string]interface{})
	for key := range cc.QueryParams() {
		tmpRequestQuery[key] = cc.QueryParam(key)
	}

	//*set param value
	tmpRequestParam := make(map[string]interface{})
	for _, value := range cc.ParamNames() {
		tmpRequestParam[value] = cc.Param(value)
	}

	// write
	wrapper.Request.Set("param", tmpRequestParam)
	wrapper.Request.Set("header", tmpRequestHeader)
	wrapper.Request.Set("body", tmpRequestBody)
	wrapper.Request.Set("query", tmpRequestQuery)

	cc.MapWrapper.Set(aliasName, wrapper)

	return nil
}

// ProcessingRequest is the core function to process every configure. doCommand for transformation, send and receive request happen here.
func ProcessingRequest(aliasName string, cc *model.CustomContext, wrapper *model.Wrapper, loopIndex int) (int, *model.CustomResponse, error) {
	// In case user want to log before modify/changing request
	if len(wrapper.Configure.Request.LogBeforeModify) > 0 {
		logValue := make(map[string]interface{}) // value to be logged
		for key, val := range wrapper.Configure.Request.LogBeforeModify {
			logValue[key] = service.GetFromHalfReferenceValue(val, wrapper.Request, loopIndex)
		}
		util.DoLoggingJson(logValue, "before", aliasName, true)
	}

	// Do the Map Modification
	tmpMapRequest := service.DoAddModifyDelete(wrapper.Configure.Request, &wrapper.Request, cc.MapWrapper, loopIndex)

	//write
	wrapper.Request.Set("header", tmpMapRequest["header"])
	wrapper.Request.Set("body", tmpMapRequest["body"])
	wrapper.Request.Set("query", tmpMapRequest["query"])

	cc.MapWrapper.Set(aliasName, wrapper)

	//get the destinationPath value before sending request
	wrapper.Configure.Request.DestinationPath = service.ModifyPath(wrapper.Configure.Request.DestinationPath, "--", cc.MapWrapper, loopIndex)

	// In case user want to log after modify/changing request
	if len(wrapper.Configure.Request.LogAfterModify) > 0 {
		logValue := make(map[string]interface{}) // value to be logged
		for key, val := range wrapper.Configure.Request.LogAfterModify {
			logValue[key] = service.GetFromHalfReferenceValue(val, wrapper.Request, loopIndex)
		}
		util.DoLoggingJson(logValue, "after", aliasName, true)
	}

	//*send to destination url
	response, err := service.Send(wrapper, cc.DefinedRoute.ProjectDirectory)

	if err != nil {
		CustomPrometheus.PromMapCounter[CustomPrometheus.Prefix+cc.DefinedRoute.ProjectDirectory+"ERR_SENDING_REQUEST"].Inc()
		log.Error("Error send : ", err.Error())
		log.Error("Set status to bad request : ", http.StatusBadRequest)
		wrapper.Response.Set("statusCode", http.StatusBadRequest)
		wrapper.Response.Set("body", map[string]interface{}{
			"error": err.Error(),
		})
		return http.StatusBadRequest, nil, err
	}
	// close http
	defer response.Body.Close()

	CustomPrometheus.PromMapCounter[CustomPrometheus.Prefix+cc.DefinedRoute.ProjectDirectory+"SUCCESS_SENDING_REQUEST"].Inc()

	// Modify responseByte in Receiver and get  byte from response that has been modified
	var tmpResponse map[string]interface{}
	tmpResponse, err = responseEntity.Receiver(wrapper.Configure, response, cc.DefinedRoute.ProjectDirectory)

	wrapper.Response.Set("statusCode", tmpResponse["statusCode"])
	wrapper.Response.Set("header", tmpResponse["header"])

	if !reflect.ValueOf(tmpResponse["body"]).IsNil() {
		wrapper.Response.Set("body", tmpResponse["body"])
	}

	// In case user want to log before modify/changing request
	if len(wrapper.Configure.Response.LogBeforeModify) > 0 {
		logValue := make(map[string]interface{}) // value to be logged
		for key, val := range wrapper.Configure.Response.LogBeforeModify {
			logValue[key] = service.GetFromHalfReferenceValue(val, wrapper.Response, loopIndex)
		}
		util.DoLoggingJson(wrapper.Configure.Response.LogBeforeModify, "before", aliasName, false)
	}

	// Do Command Add, Modify, Deletion for response again
	tmpMapResponseModified := service.DoAddModifyDelete(wrapper.Configure.Response, &wrapper.Response, cc.MapWrapper, loopIndex)

	if wrapper.Configure.Response.StatusCode > 0 {
		tmpMapResponseModified["statusCode"] = wrapper.Configure.Response.StatusCode
	} else {
		tmpMapResponseModified["statusCode"] = response.StatusCode
	}

	wrapper.Response.Set("header", tmpMapResponseModified["header"])
	wrapper.Response.Set("body", tmpMapResponseModified["body"])

	// In case user want to log after modify/changing request
	if len(wrapper.Configure.Response.LogAfterModify) > 0 {
		logValue := make(map[string]interface{}) // value to be logged
		for key, val := range wrapper.Configure.Response.LogAfterModify {
			logValue[key] = service.GetFromHalfReferenceValue(val, wrapper.Response, loopIndex)
		}
		util.DoLoggingJson(logValue, "after", aliasName, false)
	}

	cc.MapWrapper.Set(aliasName, wrapper)

	customResponse := model.CustomResponse{
		StatusCode: tmpMapResponseModified["statusCode"].(int),
		Header:     tmpMapResponseModified["header"].(map[string]interface{}),
		Body:       tmpMapResponseModified["body"].(map[string]interface{}),
		Error:      nil,
	}
	return http.StatusOK, &customResponse, nil
}
