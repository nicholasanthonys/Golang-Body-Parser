package service

import (
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

func Receiver(configure model.Configure, res *http.Response, method string, arrRes []map[string]interface{}) (map[string]interface{}, error) {
	//*read response body as byte
	responseByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Warn("Error read body")
		return nil, err
	}

	//*get response content type
	contentType := res.Header.Get("Content-Type")
	//logrus.Warn("content type response is")
	//logrus.Warn(contentType)

	//* get transform command
	transform := configure.Response.Transform
	//logrus.Warn("response byte is ")
	//logrus.Warn(string(responseByte))

	resMap := model.Fields{
		Header: make(map[string]interface{}),
		Body:   make(map[string]interface{}),
		Query:  make(map[string]interface{}),
	}

	//*get header value
	for key, val := range res.Header {
		resMap.Header[key] = val
	}

	//switch case transform
	switch transform {
	case "ToJson", "ToXml":

		//* load function transform from plugin based on transform command
		//transformFunction := LoadFunctionFromModule(transform)

		//* create empty map

		//*check content type response
		logrus.Info("content type is ", contentType)
		if len(contentType) > 0 {
			//* if content type contain application/json
			if strings.Contains(contentType, "application/json") {

				//* assign map resMap with response []byte based on response content type
				resMap.Body, _ = FromJson(responseByte)
				logrus.Warn("Content type contain application json")

			} else if strings.Contains(contentType, "application/xml") {
				//* if content type contain application/xml
				logrus.Warn("application contain xml")

				//* assign map resMap with response []byte based on response content type
				resMap.Body, _ = FromXmL(responseByte)
				logrus.Warn("resmap is")
				logrus.Warn(resMap)
			} else if strings.Contains(contentType, "text/plain") {
				//* if content type contain text/plain
				resMap.Body["message"] = string(responseByte)
			} else {
				//* panic  if content type unknown
			}

			logrus.Warn("configure repsonse adds")
			logrus.Warn(configure.Response.Adds)

			//* response always do Command
			DoCommand(configure.Response, resMap, arrRes)

			logrus.Warn("resmap after modify is")
			logrus.Warn(resMap)
			return resMap.Body, nil

			////*transform resMap BODy that has been modified to byte json or byte xml (depend on the transform command)
			//resultByte, err := transformFunction(resMap.Body)
			//
			//if err != nil {
			//	logrus.Warn("error after transform function in receiver ")
			//	logrus.Fatal(err.Error())
			//	return nil, err
			//}
			//
			////return resultByte from modified resMap
			//return resultByte, nil
		}
		return nil, nil
	default:
		return nil, nil
	}

}
