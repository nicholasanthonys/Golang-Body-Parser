package service

import (
	"fmt"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

func Send(requestFromUser *model.Wrapper) (*http.Response, error) {

	//*get transform command
	transformRequest := requestFromUser.Configure.Request.Transform

	//* constructing body to send
	body, err := Transform(requestFromUser.Configure, requestFromUser.Request.Body)

	if err != nil {
		logrus.Warn("error constructing body to send")
		return nil, err
	}

	//*get url and append it with destination path
	url := requestFromUser.Configure.Request.DestinationUrl + requestFromUser.Configure.Request.DestinationPath

	logrus.Info("sending request to url :  ", url)

	//*declare request
	var req *http.Request

	//*constructing request
	req, _ = http.NewRequest(requestFromUser.Configure.Request.Method, url, body)

	//*set Header
	SetHeader(*requestFromUser, &req.Header)

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
	switch strings.ToLower(transformRequest) {
	case strings.ToLower("ToJson"):
		header.Add("Content-Type", "application/json")
	case strings.ToLower("ToXml"):
		header.Add("Content-Type", "application/xml")
	case strings.ToLower("ToForm"):
		header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

}

func SetHeader(requestFromUser model.Wrapper, header *http.Header) {
	//actually set the header based on map header
	for key, value := range requestFromUser.Request.Header {
		if key != "Content-Type" {

			vt := reflect.TypeOf(value).Kind()
			if vt == reflect.String {
				header.Add(key, fmt.Sprintf("%s", value))
			}
		}

	}
}

func setQuery(wrapper model.Wrapper, q *url.Values) {
	//* Add
	for key, value := range wrapper.Request.Query {
		vt := reflect.TypeOf(value).Kind()
		if vt == reflect.String {
			q.Set(key, fmt.Sprintf("%v", value))
		}
	}
}
