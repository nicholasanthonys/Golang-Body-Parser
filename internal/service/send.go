package service

import (
	"fmt"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"reflect"
)

func Send(configure model.Configure, requestFromUser model.Wrapper, method string, arrRes []map[string]interface{}) (map[string]interface{}, error) {

	//*get transform command
	transformRequest := configure.Request.Transform

	//* constructing body to send
	body, err := TransformBody(configure, requestFromUser.Request.Body)

	if err != nil {
		logrus.Warn("error constructing body to send")
		return nil, err
	}
	//*kalau body nil ? masih harus di handle

	//*get method

	//*get url
	url := configure.Request.DestinationUrl
	//*declare request
	var req *http.Request

	//*constructing request
	req, _ = http.NewRequest(method, url, body)

	//*set Header
	setHeader(requestFromUser, &req.Header)

	q := req.URL.Query()

	//*set query
	setQuery(requestFromUser, &q)
	req.URL.RawQuery = q.Encode()

	// set content type for header
	setContentTypeHeader(transformRequest, &req.Header)

	return doRequest(req, configure, method, requestFromUser.Response, arrRes)
}

func doRequest(req *http.Request, configure model.Configure, method string, requestFromUserResponse model.Fields, arrRes []map[string]interface{}) (map[string]interface{}, error) {

	//* do request
	client := http.Client{}
	res, err := client.Do(req)
	defer res.Body.Close()

	if err != nil {
		logrus.Warn("Error response")
		logrus.Warn(err.Error())
		return nil, err
	}

	//*Modifty responseByte in Receiver and get  byte from response that has been modified
	receiverMap, err := Receiver(configure, res, requestFromUserResponse, arrRes)

	if err != nil {
		return nil, err
	} else {
		logrus.Warn("result byte after receive rmodify is")
		logrus.Warn((receiverMap))
	}

	//* return the receiver byte that has been modified
	return receiverMap, nil
}

func setContentTypeHeader(transformRequest string, header *http.Header) {
	//*set content type header based on transformRequest
	switch transformRequest {
	case "ToJson":

		header.Add("Content-Type", "application/json")
	case "ToXml":
		header.Add("Content-Type", "application/xml")
	case "ToForm":
		header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

}

func setHeader(requestFromUser model.Wrapper, header *http.Header) {
	//actually set the header based on map header
	for key, value := range requestFromUser.Request.Header {
		if key != "Content-Type" {

			vt := reflect.TypeOf(value).Kind()
			if vt == reflect.Slice {
				header.Add(key, fmt.Sprintf("%v", value.([]string)[0]))
			} else {
				header.Add(key, fmt.Sprintf("%s", value))
			}

		}

	}
}

func setQuery(requestFromUser model.Wrapper, q *url.Values) {
	//* Add
	for key, value := range requestFromUser.Request.Query {
		q.Set(key, fmt.Sprintf("%v", value))
		//logrus.Info("q get is ", q.Get(key))
	}
}
