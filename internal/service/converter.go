package service

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

type User struct {
	Name  interface{}
	Email interface{}
}

//* x wwww form url encoded loop field
func FormUrlFormToMap(c echo.Context) map[string]interface{} {
	logrus.Warn("FORM URL TO MAP")
	myMap := make(map[string]interface{})
	c.Request().ParseForm()
	for key, value := range c.Request().Form { // range over map
		logrus.Info("key is ", key, " value is ", value, "length is ", len(value))

		if len(value) > 1 {
			logrus.Warn("KEY ", key, " LENGTH IS ", len(value))
			myMap[key] = value
		} else {
			myMap[key] = c.FormValue(key)
		}
	}

	return myMap
}

//* form data
func formDataToMap(c echo.Context) {

	fmt.Println("Form  is")
	c.Request().ParseMultipartForm(1024)

	fmt.Println("------------------------")
	for key, values := range c.Request().Form { // range over map
		for _, value := range values { // range over []string
			fmt.Println(key, value)
		}
	}
}
