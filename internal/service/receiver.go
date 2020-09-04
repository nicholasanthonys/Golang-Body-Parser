package service

import (
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"strings"
)

func Receiver(contentType string, configure model.Configure, responseByte []byte) ([]byte, error) {
	//* get transform command
	transform := configure.Response.Transform
	logrus.Warn("response byte is ")
	logrus.Warn(string(responseByte))

	//switch case transform
	switch transform {
	case "ToJson", "ToXml":

		//* load function transform from plugin based on transform command
		transformFunction := LoadFunctionFromModule(transform)

		//* create empty map
		resMap := model.Fields{
			Header: make(map[string]interface{}),
			Body:   make(map[string]interface{}),
			Query:  make(map[string]interface{}),
		}

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
			//*modify map for response (add,delete,modify)
			DoCommandConfigureBody(configure.Response, resMap)

			logrus.Warn("resmap after modify is")
			logrus.Warn(resMap)

			//*transform resMap BODy that has been modified to byte json or byte xml (depend on the transform command)
			resultByte, err := transformFunction(resMap.Body)

			if err != nil {
				logrus.Warn("error after transform function in receiver ")
				logrus.Fatal(err.Error())
				return nil, err
			}

			//return resultByte from modified resMap
			return resultByte, nil
		}
		return nil, nil
	default:
		return []byte("transform response " + transform + " not supported"), nil
	}

}
