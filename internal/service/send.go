package service

import (
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func Send(configure model.Configure, requestFromUser map[string]interface{}) ([]byte, error) {
	transformRequest := configure.Request.Transform
	// constructing body to send
	body, err := TransformBody(configure, requestFromUser)

	if err != nil {
		logrus.Warn("error constructing body to send")
		return nil, err
	}
	//*kalau body nil ? masih harus di handle

	method := configure.Method
	url := configure.DestinationUrl
	var req *http.Request

	req, _ = http.NewRequest(method, url, body)

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

	respByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Warn("Error read body")
		return nil, err
	}
	contentType := res.Header.Get("Content-Type")
	logrus.Warn("content type is")
	logrus.Warn(contentType)
	receiverByte, err := Receiver(contentType, configure, respByte)

	if err != nil {
		return nil, err
	} else {
		logrus.Warn("result byte after receive rmodify is")
		logrus.Warn(string(receiverByte))
	}
	return receiverByte, nil
}

func setContentTypeHeader(transformRequest string, header *http.Header) {
	switch transformRequest {
	case "ToJson":
		logrus.Warn("ToJson2 triggered")
		header.Add("Content-Type", "application/json")
	case "ToXml":
		header.Add("Content-Type", "application/xml")
	case "ToForm":
		header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
}
