package service

import (
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

func Receiver(configure model.Configure, res *http.Response, requestFromUserResponse *model.Fields) (*model.Fields, error) {

	//*read response body as byte
	responseByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Warn("Error read body")
		logrus.Info(err.Error())
		return nil, err
	}

	//*get response content type
	contentType := res.Header.Get("Content-Type")
	//* get transform command
	transform := configure.Response.Transform

	//*set header value for response
	for key, _ := range res.Header {
		requestFromUserResponse.Header[key] = res.Header.Get(key)
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
			return requestFromUserResponse, nil
		}
		return nil, nil
	default:
		return nil, nil
	}

}
