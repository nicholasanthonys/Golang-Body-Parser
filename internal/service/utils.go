package service

import (
	"fmt"
	"github.com/clbanning/mxj/x2j"
	"github.com/labstack/echo"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func GetListFolder(dirname string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fmt.Println(file.Name())
	}
	return files, nil

}

func ReadConfigure(path string) []byte {
	// Open our jsonFile
	jsonFile, err := os.Open(path)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadFile(path)
	return byteValue

}

func ResponseWriter(configure model.Configure, resultMap interface{}, c echo.Context) error {
	switch configure.Response.Transform {
	case "ToJson":
		return c.JSON(200, resultMap)
	case "ToXml":
		//newResMap := make(map[string]interface{})

		//for i, res := range resultMap {
		//	byteRes, _ := x2j.MapToXml(res)
		//	index := strconv.Itoa(i)
		//	newResMap["response"] = byteRes
		//}
		resByte, _ := x2j.MapToXml(resultMap.(map[string]interface{}))
		return c.XMLBlob(200, resByte)
	default:
		return c.JSON(404, "Type Not Supported")
	}

}

func ErrorWriter(c echo.Context, configure model.Configure, err error, status int) error {
	responseMap := make(map[string]interface{})
	responseMap["message"] = err.Error()
	switch configure.Response.Transform {
	case "ToXml":
		logrus.Warn(err.Error())
		xmlByte, _ := x2j.MapToXml(responseMap)
		return c.XMLBlob(status, xmlByte)

	default:
		logrus.Warn(err.Error())
		return c.JSON(status, responseMap)
	}
}
