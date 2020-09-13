package service

import (
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func Send(configure model.Configure, requestFromUser model.Fields) ([]byte, error) {

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
	method := configure.Method
	//*get url
	url := configure.DestinationUrl
	//*declare request
	var req *http.Request

	//*constructing request
	req, _ = http.NewRequest(method, url, body)

	//* Add, Delete, Modify Header
	DoCommandConfigureHeader(configure.Request, requestFromUser, &req.Header)

	q := req.URL.Query()
	//* Add, Delete, Modify Query
	DoCommandConfigureQuery(configure.Request, requestFromUser, &q)
	req.URL.RawQuery = q.Encode()

	// set content type for header
	setContentTypeHeader(transformRequest, &req.Header)

	return doRequest(req, configure)
}

func doRequest(req *http.Request, configure model.Configure) ([]byte, error) {

	//* do request
	client := http.Client{}
	res, err := client.Do(req)
	defer res.Body.Close()

	if err != nil {
		logrus.Warn("Error response")
		logrus.Warn(err.Error())
		return nil, err
	}

	//*read response body as byte
	respByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Warn("Error read body")
		return nil, err
	}

	//*get response content type
	contentType := res.Header.Get("Content-Type")
	//logrus.Warn("content type is")
	//logrus.Warn(contentType)

	//*Modifty responseByte in Receiver and get  byte from response that has been modified
	receiverByte, err := Receiver(contentType, configure, respByte)

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
