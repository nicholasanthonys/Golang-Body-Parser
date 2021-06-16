package service

import (
	"bytes"
	"github.com/clbanning/mxj"
	"github.com/clbanning/mxj/j2x"
	"github.com/clbanning/mxj/x2j"
	"io"
	"net/url"
	"strings"
)

func Transform(transform string, requestFromUser map[string]interface{}) (io.Reader, error) {
	var body io.Reader
	switch strings.ToLower(transform) {
	case "toxml":
		resultTransformByte, err := ToXml(requestFromUser)
		if err != nil {
			log.Errorf("error : %s", err.Error())
			return nil, err
		}

		body = bytes.NewBuffer(resultTransformByte)
		return body, err
	case "toform":
		myForm := TransformToFormUrl(requestFromUser)
		body = strings.NewReader(myForm.Encode())

		return body, nil
	default:
		// tojson
		resultTransformByte, err := ToJson(requestFromUser)
		if err != nil {
			log.Errorf("error : %s", err.Error())
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
		log.Error("Error convert map to xml")
		log.Error(err.Error())
		return nil, err
	}

	//*format xml
	xmlBeautifulByte, err := mxj.BeautifyXml(xmlByte, " ", " ")
	if err != nil {
		log.Error("Error beautify  xml")
		log.Error(err.Error())
		return nil, err
	}

	return xmlBeautifulByte, nil
}

func ToJson(myMap map[string]interface{}) ([]byte, error) {
	jsonByte, err := j2x.MapToJson(myMap)

	if err != nil {
		log.Error("Error convert map to JSON")
		log.Error(err.Error())
		return nil, err
	}
	return jsonByte, nil
}
