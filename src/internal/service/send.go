package service

import (
	"fmt"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

var netTransport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}
var netClient = &http.Client{
	Timeout:   time.Second * 10,
	Transport: netTransport,
}

func Send(requestFromUser *model.Wrapper) (*http.Response, error) {

	//*get transform command
	transformRequest := requestFromUser.Configure.Request.Transform

	//* constructing body to send
	tmpBody := make(map[string]interface{})
	if tmp, ok := requestFromUser.Request.Get("body"); ok {
		tmpBody = tmp.(map[string]interface{})
	}
	var body io.Reader = nil
	var err error

	if len(tmpBody) > 0 {
		body, err = Transform(requestFromUser.Configure, tmpBody)
		log.Info("body request  is ")
		log.Info(tmpBody)

		if err != nil {
			log.Error("error constructing body to send")
			log.Error(err.Error())
			return nil, err
		}
	}

	//*get url and append it with destination path
	url := requestFromUser.Configure.Request.DestinationUrl + requestFromUser.Configure.Request.DestinationPath

	//*declare request
	var req *http.Request

	if strings.ToLower(requestFromUser.Configure.Request.Method) == "get" {
		url = setQueryGet(requestFromUser.Request, url)
		return doGetRequest(url)
	}

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

func doGetRequest(url string) (*http.Response, error) {
	logrus.Info("sending request to url :  ", url)
	resp, err := netClient.Get(url)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return resp, nil
}

func doRequest(req *http.Request) (*http.Response, error) {

	//* do request
	logrus.Info("sending request to url :  ", req.URL)
	res, err := netClient.Do(req)
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
		header.Add("Content-Type", "application/json; charset=utf-8")
	case strings.ToLower("ToXml"):
		header.Add("Content-Type", "application/xml; charset=utf-8")
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
	header.Set("Accept", "application/json")
	header.Set("Accept-Charset", "utf-8")

}

func setQuery(mapRequest cmap.ConcurrentMap, q *url.Values) {
	//* Add
	tmpQuery := make(map[string]interface{})
	if tmp, ok := mapRequest.Get("query"); ok {
		tmpQuery = tmp.(map[string]interface{})
	}
	for key, value := range tmpQuery {
		vt := reflect.TypeOf(value).Kind()
		if vt == reflect.String {
			q.Set(key, fmt.Sprintf("%v", value))
		}
	}
}

func setQueryGet(mapRequest cmap.ConcurrentMap, url string) string {
	//* Add
	tmpQuery := make(map[string]interface{})
	if tmp, ok := mapRequest.Get("query"); ok {
		tmpQuery = tmp.(map[string]interface{})
	}

	if len(tmpQuery) > 0 {
		url += "?"
	}

	for key, value := range tmpQuery {
		vt := reflect.TypeOf(value).Kind()
		if vt == reflect.String {
			url += key + "=" + value.(string)
		}
	}

	return url
}
