package service

import (
	"fmt"
	"github.com/clbanning/mxj/x2j"
	"github.com/labstack/echo"
	"github.com/nicholasantnhonys/Golang-Body-Parser/internal/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
)

//* Find is a function that will find item in slice of string
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

//* ReadConfigure is a function that will read configure.json File.
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

//*ResponseWriter is a function that will return response
func ResponseWriter(wrapper model.Wrapper, c echo.Context) error {

	switch wrapper.Configure.Response.Transform {
	case "ToJson":
		return c.JSON(200, wrapper.Response.Body)
	case "ToXml":
		resByte, _ := x2j.MapToXml(wrapper.Response.Body)
		return c.XMLBlob(200, resByte)
	default:
		logrus.Info("type not supported")
		return c.JSON(404, "Type Not Supported")
	}
}

//*ErrorWriter is a function that will return Error Response
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

//* RemoveSquareBracketAndConvertToSlice is a function that will remove the square bracket
//* and convert values into slice
func RemoveSquareBracketAndConvertToSlice(value string, separator string) []string {
	listTraverse := make([]string, 0)
	temp := ""
	arraySplit := strings.Split(value, separator)
	for _, val := range arraySplit {
		if val != "[" {
			if val == "]" {
				//*push
				listTraverse = append(listTraverse, temp)
				temp = ""
			} else {
				//* add character to temp
				temp += val
			}
		}

	}
	return listTraverse
}

// SanitizeValue is a function that will remove the dollar sign from the value in configure.json.
// we remove the dollar sign example :  $body[user], and only pick the rest of the word ex : body[user] */
func SanitizeValue(value string) ([]string, string) {
	var destination string
	var sanitized string
	if strings.HasPrefix(value, "$body") {
		destination = "body"
		sanitized = value[5:]
	} else if strings.HasPrefix(value, "$header") {
		destination = "header"
		sanitized = value[7:]
	} else if strings.HasPrefix(value, "$query") {
		destination = "query"
		sanitized = value[6:]
	} else if strings.HasPrefix(value, "$response") {
		destination = "response"
		sanitized = value[9:]
	} else if strings.HasPrefix(value, "$path") {
		destination = "path"
		sanitized = value[5:]
	} else {
		return nil, value
	}

	//* We call this function to remove square bracket. ex : $body[user] will become :  body user (as a slice)
	return RemoveSquareBracketAndConvertToSlice(sanitized, ""), destination
}

func RemoveDollar(value string) string {
	return value[1:]
}
