package service

import (
	"bytes"
	"github.com/clbanning/mxj"
	"github.com/clbanning/mxj/j2x"
	"github.com/clbanning/mxj/x2j"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"io"
	"net/url"
	"strings"
)

func TransformBody(configure model.Configure, requestFromUser map[string]interface{}) (io.Reader, error) {
	var body io.Reader
	transformRequest := configure.Request.Transform
	switch strings.ToLower(transformRequest) {
	case strings.ToLower("ToJson"):
		resultTransformByte, err := ToJson(requestFromUser)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(resultTransformByte)
		return body, err
	case strings.ToLower("ToXML"):
		resultTransformByte, err := ToXml(requestFromUser)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(resultTransformByte)
		return body, err
	case strings.ToLower("ToForm"):
		myForm := TransformToFormUrl(requestFromUser)
		body = strings.NewReader(myForm.Encode())
		return body, nil
	default:
		resultTransformByte, err := ToJson(requestFromUser)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(resultTransformByte)
		return body, err
	}

}

func TransformToFormUrl(myMap map[string]interface{}) url.Values {
	myForm := MapToFormUrl(myMap)
	return myForm
}

func ToXml(myMap map[string]interface{}) ([]byte, error) {
	xmlByte, err := x2j.MapToXml(myMap)
	if err != nil {
		logrus.Warn("Error convert map to xml")
		logrus.Warn(err.Error())
		return nil, err
	}

	//*format xml
	xmlBeautifulByte, err := mxj.BeautifyXml(xmlByte, " ", " ")
	if err != nil {
		logrus.Warn("Error beautify  xml")
		logrus.Warn(err.Error())
		return nil, err
	}

	return xmlBeautifulByte, nil
}

func ToJson(myMap map[string]interface{}) ([]byte, error) {
	jsonByte, err := j2x.MapToJson(myMap)

	if err != nil {
		logrus.Warn("Error convert map to JSON")
		logrus.Warn(err.Error())
		return nil, err
	}
	return jsonByte, nil
}
