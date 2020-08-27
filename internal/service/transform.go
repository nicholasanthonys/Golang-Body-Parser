package service

import (
	"bytes"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"io"
	"net/url"
	"strings"
)

func TransformBody(configure model.Configure, requestFromUser map[string]interface{}) (io.Reader, error) {
	var body io.Reader
	transformRequest := configure.Request.Transform
	switch transformRequest {
	case "ToJson", "ToXml":
		resultTransformByte, err := Transform(configure, requestFromUser)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(resultTransformByte)
		return body, err
	case "ToForm":
		myForm := TransformToFormUrl(requestFromUser)
		body = strings.NewReader(myForm.Encode())
		return body, nil
	default:
		logrus.Warn("transform request to " + transformRequest + " not supported")
		// masih harus di handle atau return seperti apa
		return nil, nil
	}

}

func Transform(configure model.Configure, requestFromUser map[string]interface{}) ([]byte, error) {
	transformRequest := configure.Request.Transform
	logrus.Warn("Transform request is")
	logrus.Warn(transformRequest)
	var resultByte []byte
	var err error

	//Both Request Transform ToJson or ToXml will be parsed here
	transformFunction := LoadFunctionFromModule(transformRequest)
	//transform from map to Json or XML
	resultByte, err = transformFunction(requestFromUser)

	if err != nil {
		logrus.Warn("error after transform function")
		logrus.Fatal(err.Error())
		return nil, err
	}
	logrus.Info("request from user after transform")
	logrus.Info(string(resultByte))

	return resultByte, nil
}

func TransformToFormUrl(myMap map[string]interface{}) url.Values {
	myForm := MapToFormUrl(myMap)
	return myForm
}
