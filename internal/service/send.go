package service

import (
	"bytes"
	"encoding/json"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func Send(configure model.Configure, requestFromUser map[string]interface{}) ([]byte, error) {
	transformRequest := configure.Request.Transform
	logrus.Warn("Transform request is")
	logrus.Warn(transformRequest)
	var resultByte []byte
	var err error
	if transformRequest != "ToForm" {
		transformFunction := LoadFunctionFromModule(transformRequest)
		resultByte, err = transformFunction(requestFromUser)

		if err != nil {
			logrus.Warn("error after transform function")
			logrus.Fatal(err.Error())
			return nil, err
		}
		logrus.Warn("request from user after transform")
		resultByte, _ := json.MarshalIndent(requestFromUser, " ", " ")
		logrus.Warn(string(resultByte))
	}

	method := configure.Method
	url := configure.DestinationUrl
	var req *http.Request
	var body io.Reader

	switch configure.Request.Transform {
	case "ToJson":
		//set body
		//jsonBodyByte := MapToJson(requestFromUser)\
		logrus.Warn("To Json1 triggered")
		body = bytes.NewBuffer(resultByte)
		//req.Header.Set("Content-Type", "application/json")

	case "ToForm":
		//* set body
		myForm := MapToFormUrl(requestFromUser)
		body = strings.NewReader(myForm.Encode())
		logrus.Info("body is")
		logrus.Info(body)
	}

	logrus.Warn("exit to json 1")
	//
	//form := url2.Values{}
	//form.Add("email", "abdulsalam@mail.com")
	//form.Add("password", "123456")

	req, _ = http.NewRequest(method, url, body)

	switch transformRequest {
	case "ToJson":
		logrus.Warn("ToJson2 triggered")
		req.Header.Add("Content-Type", "application/json")
	case "ToXml":
		req.Header.Add("Content-Type", "application/xml")
	case "ToForm":
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	//* do request
	client := http.Client{}
	res, err := client.Do(req)
	defer res.Body.Close()
	contentType := res.Header.Get("Content-Type")
	logrus.Warn("content type is")
	logrus.Warn(contentType)

	if err != nil {
		logrus.Fatal("Error response")
		logrus.Fatal(err.Error())
		return nil, err
	}

	if err != nil {
		logrus.Fatal("Error response")
		logrus.Fatal(err.Error())
		return nil, err
	}

	respByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Fatal("Error read body")

	}

	resultByte, err = Receiver(contentType, configure, respByte)

	if err != nil {
		return nil, err
	} else {

		logrus.Warn("result byte after receive rmodify is")
		logrus.Warn(string(resultByte))
	}

	//*response to map
	//logrus.Info("response byte is")
	//logrus.Info(string(respByte))
	//
	//var responseMap map[string]interface{}
	//_ = json.Unmarshal(respByte, &responseMap)
	//
	//jsonResponse := MapToJson(responseMap)
	//logrus.Info("map response is")
	//
	//logrus.Info(string(jsonResponse))

	return resultByte, nil

}
