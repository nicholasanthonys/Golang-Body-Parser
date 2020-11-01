package service

import (
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

func Receiver(configure model.Configure, res *http.Response, requestFromUserResponse model.Fields, arrRes []map[string]interface{}) (map[string]interface{}, error) {
	//*read response body as byte
	responseByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Warn("Error read body")
		return nil, err
	}

	//*get response content type
	contentType := res.Header.Get("Content-Type")
	//* get transform command
	transform := configure.Response.Transform

	//resMap := model.Fields{
	//	Header: make(map[string]interface{}),
	//	Body:   make(map[string]interface{}),
	//	Query:  make(map[string]interface{}),
	//}

	//*get header value
	for key, val := range requestFromUserResponse.Header {
		requestFromUserResponse.Header[key] = val
	}

	//switch case transform
	switch transform {
	case "ToJson", "ToXml":
		//*check content type response
		if len(contentType) > 0 {
			//* if content type contain application/json
			if strings.Contains(contentType, "application/json") {

				//* assign map resMap with response []byte based on response content type
				requestFromUserResponse.Body, _ = FromJson(responseByte)

			} else if strings.Contains(contentType, "application/xml") {

				//* assign map resMap with response []byte based on response content type
				requestFromUserResponse.Body, _ = FromXmL(responseByte)

			} else if strings.Contains(contentType, "text/plain") {
				//* if content type contain text/plain
				requestFromUserResponse.Body["message"] = string(responseByte)
			} else {
				//* panic  if content type unknown
			}

			//* response always do Command
			DoCommand(nil, configure.Response, requestFromUserResponse, arrRes)

			return requestFromUserResponse.Body, nil

		}
		return nil, nil
	default:
		return nil, nil
	}

}
