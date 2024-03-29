package response

import (
	"errors"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	CustomPrometheus "github.com/nicholasanthonys/Golang-Body-Parser/internal/prometheus"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/service"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

var log = logrus.New()

func init() {
	//* init logger with timestamp
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	log.Level = util.GetLogLevelFromEnv()
}

func Receiver(configure model.Configure, res *http.Response, prefixMetricName string) (map[string]interface{}, error) {
	tmpStatusCode := 0
	tmpBody := make(map[string]interface{})
	tmpHeader := make(map[string]interface{})

	//*read response body as byte
	responseByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Error("Error read body")
		logrus.Error(err.Error())
		CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(
			prefixMetricName)+"ERR_READ_RESPONSE_BODY"].
			Inc()

		return nil, err
	}

	// status code
	tmpStatusCode = res.StatusCode

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

				tmpBody, err = service.FromJson(responseByte)
				if err != nil {
					log.Error("error convert response byte to json body")
					log.Error(err.Error())
					CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(
						prefixMetricName)+"ERR_CONVERT_RESPONSE_BYTE_TO_JSON"].Inc()

				}

			} else if strings.Contains(contentType, "application/xml") {

				//* assign map resMap with response []byte based on response content type
				tmpBody, err = service.FromXmL(responseByte)
				if err != nil {
					log.Error("error convert response byte to xml body")
					log.Error(err.Error())
					CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(
						prefixMetricName)+"ERR_CONVERT_RESPONSE_BYTE_TO_XML"].Inc()

				}

			} else if strings.Contains(contentType, "text/plain") {
				//* if content type contain text/plain
				tmpBody["response"] = string(responseByte)
			} else {
				// unknown content type
				CustomPrometheus.PromMapCounter[CustomPrometheus.GetPrefixMetricName(
					prefixMetricName)+"ERR_UNKNOWN_RESPONSE_CONTENT_TYPE"].Inc()
				return nil, errors.New("Content is are not application/json, application/xml, text/plain. content type is : " + contentType)
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
