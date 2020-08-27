package main

import (
	"github.com/clbanning/mxj"
	"github.com/clbanning/mxj/j2x"
	"github.com/clbanning/mxj/x2j"
	"github.com/sirupsen/logrus"
)

func ToXml(myMap map[string]interface{}) ([]byte, error) {
	xmlByte, err := x2j.MapToXml(myMap)
	if err != nil {
		logrus.Warn("Error convert map to xml")
		logrus.Warn(err.Error())
		return nil, err
	}

	//*format xml
	xmlBeautifulByte, err := mxj.BeautifyXml(xmlByte, " ", " ")
	if err != nil {
		logrus.Warn("Error beautify  xml")
		logrus.Warn(err.Error())
		return nil, err
	}

	return xmlBeautifulByte, nil
}

func ToJson(myMap map[string]interface{}) ([]byte, error) {
	jsonByte, err := j2x.MapToJson(myMap)

	if err != nil {
		logrus.Warn("Error convert map to JSON")
		logrus.Warn(err.Error())
		return nil, err
	}
	return jsonByte, nil
}

func main() {

}

////* form data
//func formDataToMap(c echo.Context) {
//
//	fmt.Println("Form  is")
//	c.Request().ParseMultipartForm(1024)
//
//	fmt.Println("------------------------")
//	for key, values := range c.Request().Form { // range over map
//		for _, value := range values { // range over []string
//			fmt.Println(key, value)
//		}
//	}
//}
