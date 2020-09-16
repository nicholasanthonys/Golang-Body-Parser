package service

import (
	"fmt"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"reflect"
)

func Send(configure model.Configure, requestFromUser model.Fields, method string) ([]byte, error) {

	//*get transform command
	transformRequest := configure.Request.Transform

	//* constructing body to send
	body, err := TransformBody(configure, requestFromUser.Body)

	if err != nil {
		logrus.Warn("error constructing body to send")
		return nil, err
	}
	//*kalau body nil ? masih harus di handle

	//*get method

	//*get url
	url := configure.DestinationUrl
	//*declare request
	var req *http.Request

	//*constructing request
	logrus.Info("method is ", method)
	req, _ = http.NewRequest(method, url, body)

	//*set Header
	setHeader(requestFromUser, &req.Header)

	q := req.URL.Query()

	//*set query
	setQuery(requestFromUser, &q)
	req.URL.RawQuery = q.Encode()

	// set content type for header
	setContentTypeHeader(transformRequest, &req.Header)

	return doRequest(req, configure, method)
}

func doRequest(req *http.Request, configure model.Configure, method string) ([]byte, error) {

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
	receiverByte, err := Receiver(configure, res, method)

	if err != nil {
		return nil, err
	} else {
		logrus.Warn("result byte after receive rmodify is")
		logrus.Warn(string(receiverByte))
	}

	//* return the receiver byte that has been modified
	return receiverByte, nil
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

func setHeader(requestFromUser model.Fields, header *http.Header) {
	//actually set the header based on map header
	for key, value := range requestFromUser.Header {
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

func setQuery(requestFromUser model.Fields, q *url.Values) {
	//* Add
	for key, value := range requestFromUser.Query {
		q.Set(key, fmt.Sprintf("%v", value))
		logrus.Info("q get is ", q.Get(key))
	}
}
