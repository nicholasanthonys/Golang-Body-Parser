package test

import (
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGetNetTransportFromEnv(t *testing.T) {
	strTLSHandshakeTimeout := os.Getenv("TLS_HANDSHAKE_TIMEOUT")
	strResponseHeaderTimeout := os.Getenv("RESPONSE_HEADER_TIMEOUT")
	strExpectContinueTimeout := os.Getenv("EXPECT_CONTINUE_TIMEOUT")
	strDialTimeout := os.Getenv("DIAL_TIMEOUT")

	TLSHandshakeTimeout := 0
	ResponseHeaderTimeout := 0
	ExpectContinueTimeout := 0
	DialTimeout := 0

	if len(strTLSHandshakeTimeout) > 0 {
		intTLSHandshakeTimeout, err := strconv.Atoi(strTLSHandshakeTimeout)
		if err != nil {
			assert.Error(t, err, "should not error")
		} else {
			TLSHandshakeTimeout = intTLSHandshakeTimeout
		}

	}

	if len(strResponseHeaderTimeout) > 0 {
		intResponseHeaderTimeout, err := strconv.Atoi(strResponseHeaderTimeout)
		if err != nil {
			assert.Error(t, err, "should not error")
		} else {
			ResponseHeaderTimeout = intResponseHeaderTimeout
		}

	}

	if len(strExpectContinueTimeout) > 0 {
		intExpectContinueTimeout, err := strconv.Atoi(strExpectContinueTimeout)
		if err != nil {
			assert.Error(t, err, "should not error")
		} else {
			ExpectContinueTimeout = intExpectContinueTimeout
		}

	}

	if len(strDialTimeout) > 0 {
		intDialTimeout, err := strconv.Atoi(strDialTimeout)
		if err != nil {
			assert.Error(t, err, " should not error")
		} else {
			DialTimeout = intDialTimeout
		}
	}

	service.InitNet("../.env.testing")
	netTransport := service.GetNetTransportFromEnv()

	assert.Equal(t, time.Duration(TLSHandshakeTimeout)*time.Second, netTransport.TLSHandshakeTimeout)
	assert.Equal(t, time.Duration(ResponseHeaderTimeout)*time.Second, netTransport.ResponseHeaderTimeout)
	assert.Equal(t, time.Duration(ExpectContinueTimeout)*time.Second, netTransport.ExpectContinueTimeout)
	assert.Equal(t, time.Duration(DialTimeout)*time.Second, (&net.Dialer{
		Timeout: time.Duration(DialTimeout) * time.Second,
	}).Timeout)
}

func TestGetNetClientFromEnv(t *testing.T) {
	strTimeOut := os.Getenv("TIMEOUT")
	timeOut := 0
	if len(strTimeOut) > 0 {
		intTimeOut, err := strconv.Atoi(strTimeOut)
		if err != nil {
			assert.Error(t, err, "should not error")

		} else {
			timeOut = intTimeOut
		}
	}

	service.InitNet("../.env.testing")
	netClient := service.GetNetClientFromEnv()
	assert.Equal(t, netClient.Timeout, time.Duration(timeOut)*time.Second)
}

func TestSetHeader(t *testing.T) {
	// setup
	mapHeader := map[string]interface{}{
		"Content-Type":   "application/zzz", // invalid content type
		"Accept":         "application/aaa", // invalid accept
		"Accept-Charset": "utf-999",         // invalid Accept-Charset
		"address": map[string]interface{}{ // invalid type
			"street": "123 abbey road",
		},
		"api-key":           "123-456-789",
		"favourite_numbers": 777,
		"unlucky_numbers":   []int{1, 2, 3, 4},           // invalid type
		"foods":             []string{"bread", "noodle"}, // invalid type
	}

	cMapHeader := cmap.New()
	cMapHeader.Set("header", mapHeader)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))

	service.SetHeader(cMapHeader, &req.Header)

	// expected
	expected := map[string]interface{}{
		"Accept":            "application/json",
		"Accept-Charset":    "utf-8",
		"Api-Key":           "123-456-789",
		"Favourite_numbers": "777",
	}

	// actual
	actualHeader := make(map[string]interface{})
	for key, _ := range req.Header {
		actualHeader[key] = req.Header.Get(key)
	}

	// assertion
	assert.Equal(t, expected, actualHeader, " should be equal ")
}

func TestEmptySetHeader(t *testing.T) {
	// setup
	mapHeader := map[string]interface{}{}

	cMapHeader := cmap.New()
	cMapHeader.Set("header", mapHeader)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))

	service.SetHeader(cMapHeader, &req.Header)

	// expected
	expected := map[string]interface{}{
		"Accept":         "application/json",
		"Accept-Charset": "utf-8",
	}

	// actual
	actualHeader := make(map[string]interface{})
	for key, _ := range req.Header {
		actualHeader[key] = req.Header.Get(key)
	}

	// assertion
	assert.Equal(t, expected, actualHeader, " should be equal ")
}

func TestSetContentTypeHeader(t *testing.T) {
	// setup
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))

	// assertion
	transformRequest := "tojson" // lower case
	expected := "application/json; charset=utf-8"
	service.SetContentTypeHeader(transformRequest, &req.Header)
	assert.Equal(t, expected, req.Header.Get("Content-Type"), " should be equal")

	transformRequest = "toxml" // lower case
	expected = "application/xml; charset=utf-8"
	service.SetContentTypeHeader(transformRequest, &req.Header)
	assert.Equal(t, expected, req.Header.Get("Content-Type"), " should be equal")

	transformRequest = "toform" // lower case
	expected = "application/x-www-form-urlencoded; charset=utf-8"
	service.SetContentTypeHeader(transformRequest, &req.Header)
	assert.Equal(t, expected, req.Header.Get("Content-Type"), " should be equal")

	transformRequest = "toJson" // camel case
	expected = "application/json; charset=utf-8"
	service.SetContentTypeHeader(transformRequest, &req.Header)
	assert.Equal(t, expected, req.Header.Get("Content-Type"), " should be equal")

	transformRequest = "toXml" // camel case
	expected = "application/xml; charset=utf-8"
	service.SetContentTypeHeader(transformRequest, &req.Header)
	assert.Equal(t, expected, req.Header.Get("Content-Type"), " should be equal")

	transformRequest = "toForm" // camel case
	expected = "application/x-www-form-urlencoded; charset=utf-8"
	service.SetContentTypeHeader(transformRequest, &req.Header)
	assert.Equal(t, expected, req.Header.Get("Content-Type"), " should be equal")

	transformRequest = "ToJson" // correct format
	expected = "application/json; charset=utf-8"
	service.SetContentTypeHeader(transformRequest, &req.Header)
	assert.Equal(t, expected, req.Header.Get("Content-Type"), " should be equal")

	transformRequest = "ToXml" // correct format
	expected = "application/xml; charset=utf-8"
	service.SetContentTypeHeader(transformRequest, &req.Header)
	assert.Equal(t, expected, req.Header.Get("Content-Type"), " should be equal")

	transformRequest = "ToForm" // correct format
	expected = "application/x-www-form-urlencoded; charset=utf-8"
	service.SetContentTypeHeader(transformRequest, &req.Header)
	assert.Equal(t, expected, req.Header.Get("Content-Type"), " should be equal")

	transformRequest = "abczzz"                  // incorrect format
	expected = "application/json; charset=utf-8" // default
	log.Info("transform request is ", transformRequest)
	service.SetContentTypeHeader(transformRequest, &req.Header)
	assert.Equal(t, expected, req.Header.Get("Content-Type"))

	transformRequest = "toxmlzzz"                // incorrect format
	expected = "application/json; charset=utf-8" // default
	log.Info("transform request is ", transformRequest)
	service.SetContentTypeHeader(transformRequest, &req.Header)
	assert.Equal(t, expected, req.Header.Get("Content-Type"))

	transformRequest = ""                        // incorrect format
	expected = "application/json; charset=utf-8" // default
	log.Info("transform request is ", transformRequest)
	service.SetContentTypeHeader(transformRequest, &req.Header)
	assert.Equal(t, expected, req.Header.Get("Content-Type"))

}

func TestSetQuery(t *testing.T) {
	// setup
	mapQuery := map[string]interface{}{
		"address": map[string]interface{}{ // invalid type
			"street": "123 abbey road",
		},
		"api-key":           "123-456-789",
		"favourite_numbers": 777,
		"filter":            "on",
		"unlucky_numbers":   []int{1, 2, 3, 4},           // invalid type
		"foods":             []string{"bread", "noodle"}, // invalid type
	}

	cMapQuery := cmap.New()
	cMapQuery.Set("query", mapQuery)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	q := req.URL.Query()

	service.SetQuery(cMapQuery, &q)

	actualQuery := make(map[string]interface{})
	for key, _ := range q {
		actualQuery[key] = q.Get(key)
	}

	expect := map[string]interface{}{
		"api-key":           "123-456-789",
		"favourite_numbers": "777",
		"filter":            "on",
	}

	// assertion
	assert.Equal(t, expect, actualQuery, " should be equal ")

}

func TestEmptySetQuery(t *testing.T) {
	// setup
	mapQuery := map[string]interface{}{}

	cMapQuery := cmap.New()
	cMapQuery.Set("query", mapQuery)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	q := req.URL.Query()

	service.SetQuery(cMapQuery, &q)

	actualQuery := make(map[string]interface{})
	for key, _ := range q {
		actualQuery[key] = q.Get(key)
	}

	expect := map[string]interface{}{}

	// assertion
	assert.Equal(t, expect, actualQuery, " should be equal ")

}

func TestDoGetRequest(t *testing.T) {
	// setup
	url := "https://jsonplaceholder.typicode.com/todos/1"

	res, err := service.DoGetRequest(url)
	if err != nil {
		assert.Error(t, err, " should not error in TestDoGetRequest ")
	}

	resByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		assert.Error(t, err, " should not error when read response body")
	}

	expected := `{
	  "userId": 1,
	  "id": 1,
	  "title": "delectus aut autem",
	  "completed": false
	}`

	equal, err := util.JSONBytesEqual([]byte(expected), resByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}

	if !equal {
		assert.Equal(t, expected, string(resByte), "should be equal")
	}

}

func TestSendPost(t *testing.T) {
	// setup
	wrapper := model.Wrapper{
		Configure: model.Configure{
			ListStatusCodeSuccess: nil,
			Request: model.Command{
				DestinationUrl: "https://jsonplaceholder.typicode.com/posts",
				Transform:      "ToJson",
				Method:         "POST",
			},
			Response: model.Command{},
		},
		Request:  cmap.New(),
		Response: cmap.New(),
	}

	mapBody := map[string]interface{}{
		"title":  "foo",
		"body":   "bar",
		"userId": 1,
	}
	wrapper.Request.Set("body", mapBody)

	res, err := service.Send(&wrapper)
	if err != nil {
		assert.Error(t, err, " error when calling Send. should not error ")
	}

	bodyByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		assert.Error(t, err, " error when calling read response body. should not error ")
	}

	expect := `{
		"title": "foo",
  		"body": "bar",
  		"userId": 1,
		"id": 101
	}`
	equal, err := util.JSONBytesEqual([]byte(expect), bodyByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expect, string(bodyByte), "should be equal")
	}

}

func TestSendPut(t *testing.T) {
	// setup
	wrapper := model.Wrapper{
		Configure: model.Configure{
			ListStatusCodeSuccess: nil,
			Request: model.Command{
				DestinationUrl: "https://jsonplaceholder.typicode.com/posts/1",
				Transform:      "ToJson",
				Method:         "PUT",
			},
			Response: model.Command{},
		},
		Request:  cmap.New(),
		Response: cmap.New(),
	}

	mapBody := map[string]interface{}{
		"id":     1,
		"title":  "foo",
		"body":   "bar",
		"userId": 1,
	}
	wrapper.Request.Set("body", mapBody)

	res, err := service.Send(&wrapper)
	if err != nil {
		assert.Error(t, err, " error when calling Send. should not error ")
	}

	bodyByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		assert.Error(t, err, " error when calling read response body. should not error ")
	}

	expect := `{
		"title": "foo",
  		"body": "bar",
  		"userId": 1,
		"id": 1
	}`
	equal, err := util.JSONBytesEqual([]byte(expect), bodyByte)
	if err != nil {
		assert.Error(t, err, "error compare json byte")
	}
	if !equal {
		assert.Equal(t, expect, string(bodyByte), "should be equal")
	}

}
