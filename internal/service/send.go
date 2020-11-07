package service

import (
	"fmt"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"reflect"
)

func Send(configure model.Configure, requestFromUser *model.Wrapper, method string) (*http.Response, error) {

	//*get transform command
	transformRequest := configure.Request.Transform

	//* constructing body to send
	body, err := TransformBody(configure, requestFromUser.Request.Body)

	if err != nil {
		logrus.Warn("error constructing body to send")
		return nil, err
	}
	//*kalau body nil ? masih harus di handle

	//*get url and append it with destination path
	url := configure.Request.DestinationUrl + configure.Request.DestinationPath

	logrus.Info("url is ", url)

	//*declare request
	var req *http.Request

	//*constructing request
	req, _ = http.NewRequest(method, url, body)

	//*set Header
	setHeader(*requestFromUser, &req.Header)

	q := req.URL.Query()

	//*set query
	setQuery(*requestFromUser, &q)
	req.URL.RawQuery = q.Encode()

	// set content type for header
	setContentTypeHeader(transformRequest, &req.Header)

	return doRequest(req)
}

func doRequest(req *http.Request) (*http.Response, error) {
	//* do request
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		logrus.Warn("Error response")
		logrus.Warn(err.Error())
		return nil, err
	}

	return res, nil
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

func setQuery(wrapper model.Wrapper, q *url.Values) {
	//* Add
	for key, value := range wrapper.Request.Query {
		q.Set(key, fmt.Sprintf("%v", value))
		//logrus.Info("q get is ", q.Get(key))
	}
}
