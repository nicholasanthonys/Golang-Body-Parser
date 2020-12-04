package service

import (
	"fmt"
	"github.com/clbanning/mxj"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"net/url"
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
		form.Add(key, fmt.Sprintf("%v", value))
	}
	return form
}
