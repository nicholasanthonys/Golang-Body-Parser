package service

import (
	"fmt"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	cmap "github.com/orcaman/concurrent-map"
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
	tmpBody := make(map[string]interface{})
	if tmp, ok := requestFromUser.Request.Get("body"); ok {
		tmpBody = tmp.(map[string]interface{})
	}
	body, err := Transform(requestFromUser.Configure, tmpBody)

	if err != nil {
		log.Error("error constructing body to send")
		log.Error(err.Error())
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
	SetHeader(requestFromUser.Request, &req.Header)

	q := req.URL.Query()

	//*set query
	setQuery(requestFromUser.Request, &q)
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
		log.Error("Error response")
		log.Error(err.Error())
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

func SetHeader(mapRequest cmap.ConcurrentMap, header *http.Header) {
	//actually set the header based on map header
	tmpHeader := make(map[string]interface{})
	if tmp, ok := mapRequest.Get("header"); ok {
		tmpHeader = tmp.(map[string]interface{})
	}
	for key, value := range tmpHeader {
		if key != "Content-Type" {

			vt := reflect.TypeOf(value).Kind()
			if vt == reflect.String {
				header.Add(key, fmt.Sprintf("%s", value))
			}
		}

	}
}

func setQuery(mapRequest cmap.ConcurrentMap, q *url.Values) {
	//* Add
	tmpQuery := make(map[string]interface{})
	if tmp, ok := mapRequest.Get("header"); ok {
		tmpQuery = tmp.(map[string]interface{})
	}
	for key, value := range tmpQuery {
		vt := reflect.TypeOf(value).Kind()
		if vt == reflect.String {
			q.Set(key, fmt.Sprintf("%v", value))
		}
	}
}
