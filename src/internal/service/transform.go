package service

import (
	"bytes"
	"github.com/clbanning/mxj"
	"github.com/clbanning/mxj/j2x"
	"github.com/clbanning/mxj/x2j"
	CustomPrometheus "github.com/nicholasanthonys/Golang-Body-Parser/internal/prometheus"
	"io"
	"net/url"
	"strings"
)

func Transform(transform string, requestFromUser map[string]interface{}, prefixMetricName string) (io.Reader, error) {
	var body io.Reader
	switch strings.ToLower(transform) {
	case strings.ToLower("ToJson"):
		resultTransformByte, err := ToJson(requestFromUser)
		if err != nil {
			log.Errorf("error : %s", err.Error())
			CustomPrometheus.PromMapCounter[CustomPrometheus.Prefix+prefixMetricName+"ERR_TRANSFORM_REQUEST_TO_FORM"].Inc()
			return nil, err
		}
		body = bytes.NewBuffer(resultTransformByte)
		return body, err
	case strings.ToLower("ToXML"):
		resultTransformByte, err := ToXml(requestFromUser)
		if err != nil {
			log.Errorf("error : %s", err.Error())
			CustomPrometheus.PromMapCounter[CustomPrometheus.Prefix+prefixMetricName+"ERR_TRANSFORM_REQUEST_TO_XML"].Inc()
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
			log.Errorf("error : %s", err.Error())
			CustomPrometheus.PromMapCounter[CustomPrometheus.Prefix+prefixMetricName+"ERR_TRANSFORM_REQUEST_TO_FORM"].Inc()
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
