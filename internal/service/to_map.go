package service

import (
	"fmt"
	"net/url"

	"github.com/clbanning/mxj"
	"github.com/clbanning/mxj/j2x"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

// FromFormUrl is a function that transform formUrl into map string interface
func FromFormUrl(c echo.Context) map[string]interface{} {
	myMap := make(map[string]interface{})
	c.Request().ParseForm()
	for key, value := range c.Request().Form { // range over map
		if len(value) > 1 {
			myMap[key] = value
		} else {
			myMap[key] = c.FormValue(key)
		}
	}
	return myMap
}

func FromJson(byteVal []byte) (map[string]interface{}, error) {

	myMap, err := mxj.NewMapJson(byteVal)
	if err != nil {
		logrus.Warn("error")
		logrus.Warn(err.Error())
		return nil, err
	}
	return myMap, nil
}

func FromXmL(byteVal []byte) (map[string]interface{}, error) {
	myMap, err := mxj.NewMapXml(byteVal)
	if err != nil {
		logrus.Warn("error")
		logrus.Warn(err.Error())
		return nil, err
	}
	return myMap, nil
}

func MapToFormUrl(myMap map[string]interface{}) url.Values {
	form := url.Values{}
	for key, value := range myMap {
		logrus.Info("key ", key, " value ", value)
		form.Add(key, fmt.Sprintf("%v", value))
	}
	logrus.Info("form now is")
	logrus.Info(form)
	return form
}

func MapToJson(myMap map[string]interface{}) []byte {

	jsonBody, _ := j2x.MapToJson(myMap)
	//jsonBody, _ := json.MarshalIndent(myMap, "", "\t")
	return jsonBody
}
