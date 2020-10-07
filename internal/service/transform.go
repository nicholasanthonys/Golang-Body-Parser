package service

import (
	"bytes"
	"fmt"
	"github.com/clbanning/mxj/x2j"
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

	var resultByte []byte
	var err error

	//Both Request Transform ToJson or ToXml will be parsed here
	transformFunction := LoadFunctionFromModule(transformRequest)
	//transform from map to Json or XML
	resultByte, err = transformFunction(requestFromUser)

	if err != nil {
		logrus.Warn("error after transform function in service transform")
		logrus.Fatal(err.Error())
		return nil, err
	}

	return resultByte, nil
}

func TransformToFormUrl(myMap map[string]interface{}) url.Values {
	myForm := MapToFormUrl(myMap)
	return myForm
}

func TransformMapToByte(configure model.Configure, resMap map[string]interface{}) ([]byte, error) {
	//*return response

	var err error
	transformFunction := LoadFunctionFromModule(configure.Response.Transform)
	transformResultByte, err := transformFunction(resMap)

	if err != nil {
		logrus.Warn("error after transform function in receiver ")
		logrus.Fatal(err.Error())
		return nil, err
	}
	return transformResultByte, err
}

func TransformMapToArrByte(resMap map[string]interface{}) [][]byte {
	myslice := make([]interface{}, 0)
	for _, val := range resMap {
		myslice = append(myslice, val)
	}

	output := make([][]byte, 2)
	for i, v := range myslice {
		output[i] = PassInterface(v)
	}

	return output
}

func PassInterface(v interface{}) []byte {
	b, ok := x2j.MapToXml(v.(map[string]interface{}))

	fmt.Println(ok)
	fmt.Println(b)

	return b
}
