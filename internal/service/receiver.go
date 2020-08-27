package service

import (
	"encoding/xml"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"strings"
)

func Receiver(contentType string, configure model.Configure, responseByte []byte) ([]byte, error) {
	transform := configure.Response.Transform
	if transform == "ToJson" || transform == "ToXml" {
		transformFunction := LoadFunctionFromModule(transform)
		resMap := make(map[string]interface{})
		logrus.Info("content type is ", contentType)
		if len(contentType) > 0 {

			if strings.Contains(contentType, "application/json") {
				resMap, _ = FromJson(responseByte)
				logrus.Warn("Content type contain application json")

			} else {
				//*xml
				logrus.Warn("application contain xml")
				err := xml.Unmarshal(responseByte, resMap)
				if err != nil {
					return nil, err
				}
				logrus.Warn("resmap is")

			}

			//*modify map for response
			DoCommandConfigure(configure.Response, resMap)
			logrus.Warn("resmap after modify is")
			logrus.Warn(resMap)

			//*parse to xml again
			resultByte, err := transformFunction(resMap)

			if err != nil {
				logrus.Warn("error after transform function")
				logrus.Fatal(err.Error())
				return nil, err
			}

			return resultByte, nil
		}
		return nil, nil
	}
	return []byte("transform response " + transform + " not supported"), nil

}
