package request

import (
	"errors"
	"github.com/jinzhu/copier"
	"github.com/labstack/echo"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	responseEntity "github.com/nicholasanthonys/Golang-Body-Parser/internal/model/response"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
	"net/http"
)

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
func ProcessingRequest(aliasName string, c echo.Context, wrapper *model.Wrapper, mapWrapper map[string]model.Wrapper, reqByte []byte, loopIndex int) (*model.Wrapper, int, error) {
	//*check the content type user request
	var contentType string
	var err error
	var status int
	var logValue interface{} // value to be logged

	if c.Request().Header["Content-Type"] != nil {
		contentType = c.Request().Header["Content-Type"][0]
	} else {
		contentType = "application/json"
	}

	//*convert request to map string interface based on the content type
	wrapper.Request.Body, status, err = ParseRequestBody(c, contentType, reqByte)

	if err != nil {
		return nil, status, err
	}

	//*set header value
	for key := range c.Request().Header {

		wrapper.Request.Header[key] = c.Request().Header.Get(key)
	}

	//*set query value
	for key := range c.QueryParams() {
		wrapper.Request.Query[key] = c.QueryParam(key)
	}

	//*set param value
	for _, value := range c.ParamNames() {

		wrapper.Request.Param[value] = c.Param(value)
	}

	//* In case user want to log before modify/changing request
	if len(wrapper.Configure.Request.LogBeforeModify) > 0 {
		logValue = service.RetrieveValue(wrapper.Configure.Request.LogBeforeModify, wrapper.Request, loopIndex)
		util.DoLogging(logValue, "before", aliasName, true)
	}

	//*assign first before do any add,modification,delete in case value want reference each other
	mapWrapper[aliasName] = *wrapper

	// copy wrapper
	tempWrapper := model.Wrapper{}
	copier.Copy(&tempWrapper, &wrapper)

	//* Do the Map Modification
	tempWrapper.Request = service.DoAddModifyDelete(tempWrapper.Configure.Request, tempWrapper.Request, mapWrapper, loopIndex)

	//*get the destinationPath value before sending request
	tempWrapper.Configure.Request.DestinationPath = service.ModifyPath(tempWrapper.Configure.Request.DestinationPath, "--", mapWrapper, loopIndex)

	//* In case user want to log after modify/changing request
	if len(tempWrapper.Configure.Request.LogAfterModify) > 0 {
		logValue = service.RetrieveValue(tempWrapper.Configure.Request.LogAfterModify, tempWrapper.Request, loopIndex)
		util.DoLogging(logValue, "after", aliasName, true)
	}

	//*send to destination url
	response, err := service.Send(wrapper)

	if err != nil {
		logrus.Error("Error send : ", err.Error())
		return nil, http.StatusInternalServerError, err
	}

	//*Modify responseByte in Receiver and get  byte from response that has been modified
	_, err = responseEntity.Receiver(tempWrapper.Configure, response, &wrapper.Response)

	//*close http
	defer response.Body.Close()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	//* In case user want to log before modify/changing request
	if len(tempWrapper.Configure.Response.LogBeforeModify) > 0 {
		logValue = service.RetrieveValue(tempWrapper.Configure.Response.LogBeforeModify, tempWrapper.Response, loopIndex)
		util.DoLogging(tempWrapper.Configure.Response.LogBeforeModify, "before", aliasName, false)
	}

	//* Do Command Add, Modify, Deletion for response again
	tempWrapper.Response = service.DoAddModifyDelete(tempWrapper.Configure.Response, tempWrapper.Response, mapWrapper, loopIndex)

	//* In case user want to log after modify/changing request
	if len(tempWrapper.Configure.Response.LogAfterModify) > 0 {
		logValue = service.RetrieveValue(tempWrapper.Configure.Response.LogAfterModify, tempWrapper.Response, loopIndex)
		util.DoLogging(logValue, "after", aliasName, false)
	}
	return wrapper, http.StatusOK, nil
}
