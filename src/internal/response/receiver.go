package response

import (
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var log = logrus.New()

func init() {
	//* init logger with timestamp
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	log.Level = logrus.ErrorLevel
}

func Receiver(configure model.Configure, res *http.Response) (map[string]interface{}, error) {
	tmpStatusCode := ""
	tmpBody := make(map[string]interface{})
	tmpHeader := make(map[string]interface{})

	//*read response body as byte
	responseByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Error("Error read body")
		logrus.Error(err.Error())
		return nil, err
	}

	// status code
	tmpStatusCode = strconv.Itoa(res.StatusCode)

	//*get response content type
	contentType := res.Header.Get("Content-Type")

	//* get transform command
	transform := configure.Response.Transform

	//*set header value for response
	for key, _ := range res.Header {
		tmpHeader[key] = res.Header.Get(key)
	}

	//switch case transform
	switch transform {
	case "ToJson", "ToXml":
		//*check content type response
		if len(contentType) > 0 {
			//* if content type contain application/json
			if strings.Contains(contentType, "application/json") {
				//* assign map resMap with response []byte based on response content type
				tmpBody, _ = service.FromJson(responseByte)

			} else if strings.Contains(contentType, "application/xml") {

				//* assign map resMap with response []byte based on response content type
				tmpBody, _ = service.FromXmL(responseByte)

			} else if strings.Contains(contentType, "text/plain") {
				//* if content type contain text/plain
				tmpBody["response"] = string(responseByte)
			} else {
				//* panic  if content type unknown
			}
			return map[string]interface{}{
				"statusCode": tmpStatusCode,
				"header":     tmpHeader,
				"body":       tmpBody,
			}, nil
		}
		tmpBody, _ = service.FromJson(responseByte)
		return map[string]interface{}{
			"statusCode": tmpStatusCode,
			"header":     tmpHeader,
			"body":       tmpBody,
		}, nil
	default:
		tmpBody, _ = service.FromJson(responseByte)
		return map[string]interface{}{
			"statusCode": tmpStatusCode,
			"header":     tmpHeader,
			"body":       tmpBody,
		}, nil
	}

}
