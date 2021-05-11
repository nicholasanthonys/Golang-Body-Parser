package service

import (
	"fmt"
	"github.com/clbanning/mxj"
	"github.com/nicholasanthonys/Golang-Body-Parser/internal/model"
	"net/url"
)

// FromFormUrl is a function that transform formUrl into map string interface
func FromFormUrl(cc *model.CustomContext) map[string]interface{} {
	myMap := make(map[string]interface{})
	cc.Request().ParseForm()
	for key, value := range cc.Request().Form { // range over map
		if len(value) > 1 {
			myMap[key] = value
		} else {
			myMap[key] = cc.FormValue(key)
		}
	}
	return myMap
}

func FromJson(byteVal []byte) (map[string]interface{}, error) {

	myMap, err := mxj.NewMapJson(byteVal)
	if err != nil {
		log.Error("error to_map.go From Json")
		log.Error(err.Error())
		return nil, err
	}
	return myMap, nil
}

func FromXmL(byteVal []byte) (map[string]interface{}, error) {
	myMap, err := mxj.NewMapXml(byteVal)
	if err != nil {
		log.Error("error to_map.go FromXML")
		log.Error(err.Error())
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
